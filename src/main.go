package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
)

// Annotation struct
type Annotation struct {
	Label string  `json:"label"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// JSON struct
type JSON struct {
	Operator    string       `json:"operator"`
	Label       string       `json:"assignLabel"`
	Value1      string       `json:"value1"`
	Value2      string       `json:"value2"`
	Annotations []Annotation `json:"annotations"`
}

func main() {

	jsonFile, err := ioutil.ReadFile("requests.json")
	if err != nil {
		log.Fatal(err)
	}

	var annotations JSON
	err = json.Unmarshal(jsonFile, &annotations)
	if err != nil {
		log.Fatal(err)
	}

	assignLabel := annotations.Label
	value1 := annotations.Value1
	value2 := annotations.Value2

	first := Filter(annotations.Annotations, func(val Annotation) bool {
		return val.Label == value1
	})

	first = mergeOverlap(removeDuplicates(first))

	second := Filter(annotations.Annotations, func(val Annotation) bool {
		return val.Label == value2
	})

	second = mergeOverlap(removeDuplicates(second))

	operator := annotations.Operator
	// operator := "UNION"
	// operator := "INTERSECTION"
	// operator := "DIFFERENCE"
	// operator := "SYMMETRIC_DIFFERENCE"

	result := streamOperation(operator, assignLabel, first, second)

	fmt.Println("operator: ", operator)
	fmt.Println(value1, "      : ", first)
	fmt.Println(value2, "      : ", second)
	fmt.Println("ORIGIN  : ", result)
	fmt.Println("DUPP    : ", removeDuplicates(result))
	fmt.Println("REMOVE  : ", mergeOverlap(removeDuplicates(result)))
}

func removeDuplicates(elements []Annotation) []Annotation {
	// Use map to record duplicates as we find them.
	encountered := map[Annotation]bool{}
	result := []Annotation{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true

			if elements[v].Start == elements[v].End {
				continue
			}

			result = append(result, elements[v])
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Start < result[j].Start
	})
	return result
}

// Merge if they overlap or touch.
func mergeOverlap(elements []Annotation) []Annotation {
	if len(elements) == 1 {
		return elements
	}

	result := []Annotation{}
	buff := elements[0]
	for v := range elements {
		if v != 0 {
			current := elements[v]
			before := elements[v-1]

			if buff != before {
				before = buff
			}

			if (before.Start <= current.Start) && (current.Start <= before.End) {
				if before.End <= current.End {
					buff.Start = before.Start
					buff.End = current.End
				} else {
					buff.Start = before.Start
					buff.End = before.End
				}
			} else if before.End < current.Start {
				result = append(result, buff)
				buff = current

				if v == len(elements)-1 {
					result = append(result, current)
				}
				continue
			}
			if v == len(elements)-1 {
				result = append(result, buff)
			}
		}
	}

	return result
}

// Filter by key value
func Filter(arr []Annotation, cond func(Annotation) bool) []Annotation {
	result := []Annotation{}
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

func streamOperation(operator string, assignLabel string, first []Annotation, second []Annotation) []Annotation {
	result := []Annotation{}

	for i := 0; i < len(first); i++ {
		for j := 0; j < len(second); j++ {
			var newAnnotation Annotation
			newAnnotation.Label = assignLabel

			switch operator {
			case "UNION":
				// BEFORE
				if second[j].End <= first[i].Start {
					if i == 0 {
						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					} else {
						if second[j].Start <= first[i-1].End {
							break
						}
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)

						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					}

					if j == len(second)-1 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if second[j+1].Start <= first[i].End {
							break
						}
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)

						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					}
					break
				}

				// LEFT
				if (second[j].Start <= first[i].Start) && (first[i].Start <= second[j].End) && (second[j].End <= first[i].End) {
					newAnnotation.Start = second[j].Start
					newAnnotation.End = first[i].End
					result = append(result, newAnnotation)
					break
				}

				// IN
				if (second[j].Start <= first[i].Start) && (first[i].End <= second[j].End) {
					newAnnotation.Start = second[j].Start
					newAnnotation.End = second[j].End
					result = append(result, newAnnotation)
					break
				}

				// OUT
				if (first[i].Start <= second[j].Start) && (second[j].End <= first[i].End) {
					newAnnotation.Start = first[i].Start
					newAnnotation.End = first[i].End
					result = append(result, newAnnotation)
					break
				}

				// RIGHT
				if (first[i].Start <= second[j].Start) && (second[j].Start <= first[i].End) && (first[i].End <= second[j].End) {
					newAnnotation.Start = first[i].Start
					newAnnotation.End = second[j].End
					result = append(result, newAnnotation)
					break
				}

				// AFTER
				if first[i].End <= second[j].Start {
					if i == len(first)-1 {
						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					} else {
						if first[i+1].Start <= second[j].End {
							break
						}
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)

						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					}

					if j == 0 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
						break
					} else {
						if first[i].Start <= second[j-1].End {
							break
						}
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					}
					break
				}

			case "INTERSECTION":
				// BEFORE
				if second[j].End <= first[i].Start {
					break
				}

				// LEFT
				if (second[j].Start <= first[i].Start) && (first[i].Start <= second[j].End) && (second[j].End <= first[i].End) {
					newAnnotation.Start = first[i].Start
					newAnnotation.End = second[j].End
					result = append(result, newAnnotation)
					break
				}

				// IN
				if (second[j].Start <= first[i].Start) && (first[i].End <= second[j].End) {
					newAnnotation.Start = first[i].Start
					newAnnotation.End = first[i].End
					result = append(result, newAnnotation)
					break
				}

				// OUT
				if (first[i].Start <= second[j].Start) && (second[j].End <= first[i].End) {
					newAnnotation.Start = second[j].Start
					newAnnotation.End = second[j].End
					result = append(result, newAnnotation)
					break
				}

				// RIGHT
				if (first[i].Start <= second[j].Start) && (second[j].Start <= first[i].End) && (first[i].End <= second[j].End) {
					newAnnotation.Start = second[j].Start
					newAnnotation.End = first[i].End
					result = append(result, newAnnotation)
					break
				}

				//AFTER
				if first[i].End <= second[j].Start {
					break
				}

			case "DIFFERENCE":
				// BEFORE
				if second[j].End <= first[i].Start {
					if j == len(second)-1 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if second[j+1].Start <= first[i].End {
							break
						}
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					}
					break
				}

				// LEFT
				if (second[j].Start <= first[i].Start) && (first[i].Start <= second[j].End) && (second[j].End <= first[i].End) {
					if j < len(second)-1 {
						if second[j+1].Start <= first[i].End {
							newAnnotation.Start = second[j].End
							newAnnotation.End = second[j+1].Start
							result = append(result, newAnnotation)
						} else if first[i].End <= second[j+1].Start {
							newAnnotation.Start = second[j].End
							newAnnotation.End = first[i].End
							result = append(result, newAnnotation)
						}
					} else {
						newAnnotation.Start = second[j].End
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					}

					break
				}

				// IN
				if (second[j].Start <= first[i].Start) && (first[i].End <= second[j].End) {
					break
				}

				// OUT
				if (first[i].Start <= second[j].Start) && (second[j].End <= first[i].End) {
					if j != 0 {
						if second[j-1].End <= first[i].Start {
							newAnnotation.Start = first[i].Start
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						}
						if first[i].Start <= second[j-1].End {
							newAnnotation.Start = second[j-1].End
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						}
					} else {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = second[j].Start
						result = append(result, newAnnotation)
					}

					if j < len(second)-1 {
						if first[i].End <= second[j+1].Start {
							newAnnotation.Start = second[j].End
							newAnnotation.End = first[i].End
							result = append(result, newAnnotation)
						}
						if second[j+1].Start <= first[i].End {
							newAnnotation.Start = second[j].End
							newAnnotation.End = second[j+1].Start
							result = append(result, newAnnotation)
						}
					} else {
						newAnnotation.Start = first[i].End
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)

					}
					break
				}

				// RIGHT
				if (first[i].Start <= second[j].Start) && (second[j].Start <= first[i].End) && (first[i].End <= second[j].End) {
					if j != 0 {
						if first[i].Start <= second[j-1].End {
							newAnnotation.Start = second[j-1].End
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						} else if second[j-1].End <= first[i].Start {
							newAnnotation.Start = first[i].Start
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						}
					} else {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = second[j].Start
						result = append(result, newAnnotation)
					}
					break
				}

				// AFTER
				if first[i].End <= second[j].Start {
					if j == 0 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if first[i].Start <= second[j-1].End {
							break
						}
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					}
					break
				}

			case "SYMMETRIC_DIFFERENCE":
				// BEFORE
				if second[j].End <= first[i].Start {
					if i == 0 {
						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					} else {
						if second[j].Start <= first[i-1].End {
							break
						} else if first[i-1].End <= second[j].Start {
							newAnnotation.Start = second[j].Start
							newAnnotation.End = second[j].End
							result = append(result, newAnnotation)
						}
					}

					if j == len(second)-1 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if second[j+1].Start <= first[i].End {
							break
						} else if first[i].End <= second[j+1].Start {
							newAnnotation.Start = first[i].Start
							newAnnotation.End = first[i].End
							result = append(result, newAnnotation)
						}
					}
					break
				}

				// LEFT
				if (second[j].Start <= first[i].Start) && (first[i].Start <= second[j].End) && (second[j].End <= first[i].End) {
					if i == 0 {
						newAnnotation.Start = second[j].Start
						newAnnotation.End = first[i].Start
						result = append(result, newAnnotation)
					} else {
						if second[j].Start <= first[i-1].End {
							newAnnotation.Start = first[i-1].End
							newAnnotation.End = first[i].Start
							result = append(result, newAnnotation)
						} else if first[i-1].End <= second[j].Start {
							newAnnotation.Start = second[j].Start
							newAnnotation.End = first[i].Start
							result = append(result, newAnnotation)
						}
					}

					if j == len(second)-1 {
						newAnnotation.Start = second[j].End
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if second[j+1].Start <= first[i].End {
							newAnnotation.Start = second[j].End
							newAnnotation.End = second[j+1].Start
							result = append(result, newAnnotation)
						} else if first[i].End <= second[j+1].Start {
							newAnnotation.Start = second[j].End
							newAnnotation.End = first[i].End
							result = append(result, newAnnotation)
						}
					}
					break
				}

				// IN
				if (second[j].Start <= first[i].Start) && (first[i].End <= second[j].End) {
					if i == 0 {
						newAnnotation.Start = second[j].Start
						newAnnotation.End = first[i].Start
						result = append(result, newAnnotation)
					} else {
						if second[j].Start <= first[i-1].End {
							newAnnotation.Start = first[i-1].End
							newAnnotation.End = first[i].Start
							result = append(result, newAnnotation)
						} else if first[i-1].End <= second[j].Start {
							newAnnotation.Start = second[j].Start
							newAnnotation.End = first[i].Start
							result = append(result, newAnnotation)
						}
					}

					if i == len(first)-1 {
						newAnnotation.Start = first[i].End
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					} else {
						if first[i+1].Start <= second[j].End {
							newAnnotation.Start = first[i].End
							newAnnotation.End = first[i+1].Start
							result = append(result, newAnnotation)
						} else if second[j].End <= first[i+1].Start {
							newAnnotation.Start = first[i].End
							newAnnotation.End = second[j].End
							result = append(result, newAnnotation)
						}
					}
					break
				}

				// OUT
				if (first[i].Start <= second[j].Start) && (second[j].End <= first[i].End) {
					if j == 0 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = second[j].Start
						result = append(result, newAnnotation)
					} else {
						if first[i].Start <= second[j-1].End {
							newAnnotation.Start = second[j-1].End
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						} else if second[j-1].End <= first[i].Start {
							newAnnotation.Start = first[i].Start
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						}
					}

					if j == len(second)-1 {
						newAnnotation.Start = second[j].End
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if second[j+1].Start <= first[i].End {
							newAnnotation.Start = second[j].End
							newAnnotation.End = second[j+1].Start
							result = append(result, newAnnotation)
						} else if first[i].End <= second[j+1].Start {
							newAnnotation.Start = second[j].End
							newAnnotation.End = first[i].End
							result = append(result, newAnnotation)
						}
					}
					break
				}

				// RIGHT
				if (first[i].Start <= second[j].Start) && (second[j].Start <= first[i].End) && (first[i].End <= second[j].End) {
					if i == len(first)-1 {
						newAnnotation.Start = first[i].End
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					} else {
						if first[i+1].Start <= second[j].End {
							newAnnotation.Start = first[i].End
							newAnnotation.End = first[i+1].Start
							result = append(result, newAnnotation)
						} else if second[j].End <= first[i+1].Start {
							newAnnotation.Start = first[i].End
							newAnnotation.End = second[j].End
							result = append(result, newAnnotation)
						}
					}

					if j == 0 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = second[j].Start
						result = append(result, newAnnotation)
					} else {
						if first[i].Start <= second[j-1].End {
							newAnnotation.Start = second[j-1].End
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						} else if second[j-1].End <= first[i].Start {
							newAnnotation.Start = first[i].Start
							newAnnotation.End = second[j].Start
							result = append(result, newAnnotation)
						}
					}
					break
				}

				// AFTER
				if first[i].End <= second[j].Start {
					if i == len(first)-1 {
						newAnnotation.Start = second[j].Start
						newAnnotation.End = second[j].End
						result = append(result, newAnnotation)
					} else {
						if first[i+1].Start <= second[j].End {
							break
						} else if second[j].End <= first[i+1].Start {
							newAnnotation.Start = second[j].Start
							newAnnotation.End = second[j].End
							result = append(result, newAnnotation)
						}
					}

					if j == 0 {
						newAnnotation.Start = first[i].Start
						newAnnotation.End = first[i].End
						result = append(result, newAnnotation)
					} else {
						if first[i].Start <= second[j-1].End {
							break
						} else if second[j-1].End <= first[i].Start {
							newAnnotation.Start = first[i].Start
							newAnnotation.End = first[i].End
							result = append(result, newAnnotation)
						}
					}
					break
				}

			default:
				break
			}
		}

	}
	return result
}
