package vklp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("unmarshal", func() {
	It("should unmarshal JSON array into values", func() {
		var boolData bool
		var stringData string
		var intData int64
		var floatData float64

		input := []byte(`[ true, "string", 42, 4.2 ]`)

		err := unmarshal(input, &boolData, &stringData, &intData, &floatData)
		Expect(err).NotTo(HaveOccurred())
		Expect(boolData).To(Equal(true))
		Expect(stringData).To(Equal("string"))
		Expect(intData).To(Equal(int64(42)))
		Expect(floatData).To(Equal(float64(4.2)))
	})
})
