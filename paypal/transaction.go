package paypal

import (
	"settings"
	"fmt"
)

type Transaction struct {
	Amount *Amount `json:"amount"`
	Description string `json:"description"`
	//Payee *Payee `json:"payee"`
	InvoiceNumber string `json:"invoice_number"`
	PaymentOptions *PaymentOptions `json:"payment_options"`
	ItemList *ItemList `json:"item_list"`
}

type ItemList struct {
	Items []*Item `json:"items"`
	ShippingAddress *ShippingAddress `json:"shipping_address"`
}

type Item struct {
	Sku string `json:"sku"`
	Name string `json:"name"`
	Description string `json:"description"`
	Quantity string `json:"quantity"`
	Price string `json:"price"`
	Currency string `json:"currency"`
	Tax string `json:"tax"`
	Url string `json:"url"`
}

func NewItem(sku string, name string, description string, quantity int, priceCents int64, taxCents int64, url string) *Item {
	return &Item{
		Sku: sku,
		Name: name,
		Description: description,
		Quantity: fmt.Sprintf("%v", quantity),
		Price: fmt.Sprintf("%.2f", float64(priceCents)/100),
		Tax: fmt.Sprintf("%.2f", float64(taxCents)/100),
		Url: url,
		Currency: "USD",
	}
}

type ShippingAddress struct {
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
	City string `json:"city"`
	CountryCode string `json:"country_code"`
	PostalCode string `json:"postal_code"`
	State string `json:"state"`
	Phone string `json:"phone"`
	RecipientName string `json:"recipient_name"`
}

type Payee struct {
	Email string `json:"email"`
}

type PaymentOptions struct {
	AllowedPaymentMethod string `json:"allowed_payment_method"`
}

func NewTransaction (invoiceNumber string, description string, amount *Amount, items []*Item, shippingAddress *ShippingAddress) *Transaction {
	return &Transaction{
		Amount: amount,
		Description: description,
		//Payee: &Payee{
		//	Email: settings.PAYPAL_EMAIL,
		//},
		InvoiceNumber: invoiceNumber,
		PaymentOptions: &PaymentOptions{AllowedPaymentMethod:settings.PAYPAL_ALLOWED_PAYMENT_OPTION},
		ItemList: &ItemList{
			Items: items,
			ShippingAddress: shippingAddress,
		},
	}
}
