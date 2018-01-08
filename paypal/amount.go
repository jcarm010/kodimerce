package paypal

import "fmt"

type Amount struct {
	Total string `json:"total"`
	Currency string `json:"currency"`
	Details *AmountDetail `json:"details"`
}

type AmountDetail struct {
	Subtotal string `json:"subtotal"`
	Tax string `json:"tax"`
	Shipping string `json:"shipping"`
	HandlingFee string `json:"handling_fee"`
	ShippingDiscount string `json:"shipping_discount"`
	Insurance string `json:"insurance"`
}

func NewAmount (subtotalCents int64, taxCents int64, shippingCents int64, handlingFeeCents int64, shippingDiscountCents int64, insuranceCents int64) *Amount {
	var totalCents int64 = subtotalCents
	totalCents += taxCents
	totalCents += shippingCents
	totalCents += handlingFeeCents
	totalCents += shippingDiscountCents
	totalCents += insuranceCents
	return &Amount{
		Total: fmt.Sprintf("%.2f", float64(totalCents)/100),
		Currency: "USD",
		Details: &AmountDetail{
			Subtotal: fmt.Sprintf("%.2f", float64(subtotalCents)/100),
			Tax: fmt.Sprintf("%.2f", float64(taxCents)/100),
			Shipping: fmt.Sprintf("%.2f", float64(shippingCents)/100),
			HandlingFee: fmt.Sprintf("%.2f", float64(handlingFeeCents)/100),
			ShippingDiscount: fmt.Sprintf("%.2f", float64(shippingDiscountCents)/100),
			Insurance: fmt.Sprintf("%.2f", float64(insuranceCents)/100),
		},
	}
}
