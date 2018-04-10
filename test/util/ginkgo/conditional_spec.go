package ginkgo

import (
	"fmt"

	"github.com/onsi/ginkgo"
)

// ConditionallyIt is similar to It, but also runs a condition check in the
// deferred context of It, and fails the spec if the condition check fails.
// Ginkgo runs all of the code other than code in It blocks first and adds
// the It blocks to be run later. Because of this, something like this does
// not work:
//	Describe("" func() {
//		success := false
//		It("should check equality", func() {
//			Expect(foo()).To(Equal(bar()))
//			success = true
//		})
//
//		if success {
//			It("should check reverse equality", func() {
//				Expect(bar()).To(Equal(foo()))
//			})
//		}
//	})
//
// This doesn't work because `if success` is run immediately on the first
// sweep, so it will always be false.
// You could write:
//	Describe("" func() {
//		success := false
//		It("should check equality", func() {
//			Expect(foo()).To(Equal(bar()))
//			success = true
//		})
//
//		It("should check reverse equality", func() {
//			if !success {
//				Fail("initial equality check failed")
//			}
//			Expect(bar()).To(Equal(foo()))
//		})
//	})
//
// Or, you could use ConditionallyIt to do the work for you:
//	Describe("" func() {
// 		success := false
// 		It("should check equality", func() {
// 			Expect(foo()).To(Equal(bar()))
// 			success = true
// 		})
//
// 		ConditionallyIt(
// 			"should check reverse equality",
// 			func() {
// 				Expect(bar()).To(Equal(foo()))
// 			},
// 			If("the first equality check succeeded", func() bool { return success }),
// 		)
//	})
//
//  You cannot do the following:
func ConditionallyIt(
	description string,
	condition ConditionFunc,
	body func(),
	timeout ...float64,
) {
	conditionCheck, conditionDescription := condition()
	ginkgo.It(description, func() {
		if !conditionCheck() {
			message := "precondition failed"
			if conditionDescription != "" {
				message = fmt.Sprintf("%v: %v", message, conditionDescription)
			}
			ginkgo.Fail(message, 0)
		}
		body()
	}, timeout...)
	return
}

type ConditionFunc func() (func() bool, string)

func If(description string, condition func() bool) ConditionFunc {
	return func() (func() bool, string) {
		return condition, description
	}
}
