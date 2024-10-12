package pdf

import (
	"billify-api/internal/store"
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/jung-kurt/gofpdf/v2"
)

type PDFGenerator struct{}

func NewPDFGenerator() PDFGenerator {
	return PDFGenerator{}
}

func (p *PDFGenerator) GenerateInvoicePDF(business *store.Business, invoice *store.Invoice, customer *store.Customer, items []*store.InvoiceItem, products []*store.Product) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "./fonts")

	pdf.AddUTF8Font("Poppins", "", "./Poppins-Regular.ttf")
	pdf.AddUTF8Font("Poppins", "B", "./Poppins-Bold.ttf")
	pdf.AddUTF8Font("Poppins", "I", "./Poppins-Italic.ttf")

	// Add page
	pdf.AddPage()

	// Margins and header styling
	pdf.SetMargins(5, 5, 5)
	pdf.SetFont("Poppins", "B", 18)

	// Colors
	black := []int{0, 0, 0}
	darkBlue := []int{78, 79, 235}
	grey := []int{238, 238, 238}

	// Invoice Header
	pdf.SetTextColor(darkBlue[0], darkBlue[1], darkBlue[2])
	pdf.CellFormat(190, 10, fmt.Sprintf("Invoice #%d", invoice.InvNo), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Company Information
	pdf.SetFont("Poppins", "B", 16)
	pdf.SetTextColor(black[0], black[1], black[2])
	pdf.CellFormat(190, 7, business.Name, "", 1, "", false, 0, "")
	pdf.SetFont("Poppins", "", 8)
	pdf.MultiCell(190, 5, fmt.Sprintf("GSTIN: %s\n%s\n%s, %s, %s, %s\nPhone: %s\nEmail: %s", business.GSTNo, business.Address, business.City, business.State, business.ZipCode, business.Country, business.CompanyPhone, business.CompanyEmail), "", "", false)

	// Invoice Date / Due Date
	pdf.Ln(4)
	pdf.SetFont("Poppins", "B", 8)
	pdf.CellFormat(95, 6, fmt.Sprintf("Invoice Date: %s", invoice.InvDate.Format("02/01/2006")), "", 0, "", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Due Date: %s", invoice.DueDate.Format("02/01/2006")), "", 1, "", false, 0, "")
	pdf.Ln(5)

	// Customer Information
	pdf.SetFont("Poppins", "B", 8)
	pdf.CellFormat(95, 6, "Customer Detail:", "", 1, "", false, 0, "")
	pdf.SetFont("Poppins", "", 8)
	pdf.CellFormat(95, 5, fmt.Sprintf("Name: %s", customer.Name), "", 1, "", false, 0, "")
	pdf.CellFormat(95, 5, fmt.Sprintf("GSTNo: %s", customer.GSTNo), "", 1, "", false, 0, "")
	pdf.Ln(2)

	// Billing and Shipping Address
	pdf.SetFont("Poppins", "B", 8)
	pdf.CellFormat(100, 6, "Billing Address:", "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Shipping Address:", "", 1, "L", false, 0, "")
	pdf.SetFont("Poppins", "", 8)

	// Calculate the height of the multi-cell to ensure both columns align properly
	billingHeight := math.Ceil(pdf.GetStringWidth(customer.BAddress)/95) * 6
	shippingHeight := math.Ceil(pdf.GetStringWidth(customer.SAddress)/95) * 6

	// Billing Address
	pdf.MultiCell(95, 6, customer.BAddress, "", "L", false)
	// Move to the right column for Shipping Address
	pdf.SetXY(105, pdf.GetY()-billingHeight)
	pdf.MultiCell(95, 6, customer.SAddress, "", "L", false)
	if billingHeight > shippingHeight {
		pdf.Ln(billingHeight + 10)
	} else {
		pdf.Ln(10)
	}

	// Table Header
	pdf.SetFont("Poppins", "B", 8)
	pdf.SetFillColor(darkBlue[0], darkBlue[1], darkBlue[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(10, 7, "#", "", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Item", "", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Rate / Item", "", 0, "C", true, 0, "")
	pdf.CellFormat(15, 7, "Qty", "", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Taxable Value", "", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Tax Amount", "", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, "Item Total", "", 1, "C", true, 0, "")

	// Table Rows
	pdf.SetFont("Poppins", "", 8)
	pdf.SetTextColor(black[0], black[1], black[2])
	for i, item := range items {
		var product *store.Product
		for _, p := range products {
			if p.ID == item.ProdID {
				product = p
				break
			}
		}
		taxableValue := item.UnitPrice * float64(item.Quantity)
		taxAmount := (taxableValue * product.TaxRate) / 100
		totalAmount := taxableValue + taxAmount
		pdf.SetFillColor(grey[0], grey[1], grey[2])
		pdf.CellFormat(10, 7, strconv.Itoa(i+1), "", 0, "C", true, 0, "")
		pdf.CellFormat(60, 7, product.Name, "", 0, "", true, 0, "")
		pdf.CellFormat(25, 7, fmt.Sprintf("₹ %.2f", item.UnitPrice), "", 0, "R", true, 0, "")
		pdf.CellFormat(15, 7, strconv.Itoa(item.Quantity), "", 0, "R", true, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("₹ %.2f", taxableValue), "", 0, "R", true, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("₹ %.2f", taxAmount), "", 0, "R", true, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("₹ %.2f", totalAmount), "", 1, "R", true, 0, "")
	}
	pdf.CellFormat(200, 1, "", "B", 0, "R", false, 1, "")

	// Total Amount
	pdf.SetFont("Poppins", "B", 8)
	pdf.Ln(2) // Add a small line break to ensure separation
	pdf.CellFormat(160, 7, "Total Amount", "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 7, fmt.Sprintf("₹ %.2f", invoice.TotalAmount), "", 1, "R", false, 0, "")
	// fmt.Println(invoice.TotalAmount)
	// // Convert total amount to words
	// totalAmountWords, err := ntw.Convert(int(invoice.TotalAmount))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to convert total amount to words: %w", err)
	// }

	// // Display total amount in words
	// pdf.SetFont("Poppins", "", 8)
	// pdf.CellFormat(200, 4, fmt.Sprintf("Total Amount in Words: %s", totalAmountWords), "", 1, "R", false, 0, "")

	// Footer Bank Details
	pdf.Ln(10)
	pdf.SetFont("Poppins", "B", 8)
	pdf.CellFormat(190, 6, "Bank Details:", "", 1, "", false, 0, "")
	pdf.SetFont("Poppins", "", 8)
	pdf.MultiCell(190, 5, fmt.Sprintf("Bank: %s\nAccount No: %s\nIFSC Code: %s\nBranch: %s", business.BankName, business.AccountNo, business.IFSC, business.BankBranch), "", "", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
