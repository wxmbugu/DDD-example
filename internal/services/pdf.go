package services

import (
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
)

func init() {
	err := license.SetMeteredKey(`e9da5e6d147303d9b079887d1997b283af025fc02c5d0942bfd7a5a45524e13e`)
	if err != nil {
		panic(err)
	}
}
func NewCreator() creator.Creator {

	c := creator.New()

	return *c
}
func (s *Service) Generate(id int, w io.Writer) {
	font, err := model.NewStandard14Font("Helvetica")
	if err != nil {
		log.Fatal(err)
	}

	fontBold, err := model.NewStandard14Font("Helvetica-Bold")
	if err != nil {
		log.Fatal(err)
	}

	// Generate front page.
	drawFrontPage(&s.Creator, font, fontBold)
	// Generate basic usage chapter.
	if err := s.BasicUsage(&s.Creator, font, fontBold, id); err != nil {
		log.Fatal(err)
	}

	//
	// Write to output file.
	if err := s.Creator.Write(w); err != nil {
		log.Fatal(err)
	}

}
func drawFrontPage(c *creator.Creator, font, fontBold *model.PdfFont) {
	c.CreateFrontPage(func(args creator.FrontpageFunctionArgs) {
		p := c.NewStyledParagraph()
		p.SetMargins(0, 0, 300, 0)
		p.SetTextAlignment(creator.TextAlignmentCenter)

		chunk := p.Append("Patient Tracker System")
		chunk.Style.Font = font
		chunk.Style.FontSize = 56
		chunk.Style.Color = creator.ColorRGBFrom8bit(56, 68, 77)

		chunk = p.Append("\n")

		chunk = p.Append("Medical Report")
		chunk.Style.Font = fontBold
		chunk.Style.FontSize = 40
		chunk.Style.Color = creator.ColorRGBFrom8bit(45, 148, 215)

		c.Draw(p)
	})
}

func (s *Service) BasicUsage(c *creator.Creator, font, fontBold *model.PdfFont, patientid int) error {
	// Create chapter.
	ch := c.NewChapter("Records")
	ch.SetMargins(0, 0, 50, 0)
	ch.GetHeading().SetFont(font)
	ch.GetHeading().SetFontSize(18)
	ch.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	// Draw subchapters.
	s.ContentAlignH(c, ch, font, fontBold, patientid)

	// Draw chapter.
	if err := c.Draw(ch); err != nil {
		return err
	}

	return nil
}

func (s *Service) ContentAlignH(c *creator.Creator, ch *creator.Chapter, font, fontBold *model.PdfFont, patientrecord int) {
	// Create subchapter.
	sc := ch.NewSubchapter("Medical Records")
	sc.SetMargins(0, 0, 30, 0)
	sc.GetHeading().SetFont(font)
	sc.GetHeading().SetFontSize(13)
	sc.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	// Create subchapter description.
	desc := c.NewStyledParagraph()
	desc.SetMargins(0, 0, 10, 0)
	desc.Append("Patient Tracker Medical records")

	sc.Add(desc)

	// Create table.
	table := c.NewTable(3)
	table.SetMargins(0, 0, 10, 0)

	drawCell := func(text string, font *model.PdfFont, align creator.CellHorizontalAlignment) {
		p := c.NewStyledParagraph()
		p.Append(text).Style.Font = font

		cell := table.NewCell()
		cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(align)
		cell.SetContent(p)
	}

	// Draw table header.
	drawCell("#", fontBold, creator.CellHorizontalAlignmentLeft)
	drawCell("Medical Record", fontBold, creator.CellHorizontalAlignmentCenter)
	drawCell("Medical Value", fontBold, creator.CellHorizontalAlignmentRight)
	record, err := s.PatientRecordService.Find(patientrecord)
	if err != nil {
		return
	}
	// Draw table content.
	values := reflect.ValueOf(record)
	typesOf := values.Type()
	for i := 0; i < values.NumField(); i++ {
		fieldValue := values.Field(i)
		drawCell(fmt.Sprintf("%d", i+1), font, creator.CellHorizontalAlignmentLeft)
		drawCell(fmt.Sprintf(typesOf.Field(i).Name), font, creator.CellHorizontalAlignmentCenter)
		drawCell(fmt.Sprintf("%v", fieldValue.Interface()), font, creator.CellHorizontalAlignmentRight)
	}

	sc.Add(table)
}
