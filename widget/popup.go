package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/theme"
)

// PopUp is a widget that can float above the user interface.
// It wraps any standard elements with padding and a shadow.
// If it is modal then the shadow will cover the entire canvas it hovers over and block interactions.
type PopUp struct {
	BaseWidget

	Content fyne.CanvasObject
	Canvas  fyne.Canvas

	innerPos     fyne.Position
	innerSize    fyne.Size
	modal        bool
	overlayShown bool
}

// Hide this widget, if it was previously visible
func (p *PopUp) Hide() {
	if p.overlayShown {
		p.Canvas.Overlays().Remove(p)
		p.overlayShown = false
	}
	p.BaseWidget.Hide()
}

// Move the widget to a new position. A PopUp position is absolute to the top, left of its canvas.
// For PopUp this actually moves the content so checking Position() will not return the same value as is set here.
func (p *PopUp) Move(pos fyne.Position) {
	if p.modal {
		return
	}
	p.innerPos = pos
	p.Refresh()
}

// Resize changes the size of the PopUp's content.
// PopUps always have the size of their canvas, but this call updates the
// size of the content portion.
//
// Implements: fyne.Widget
func (p *PopUp) Resize(size fyne.Size) {
	p.innerSize = size
	// The canvas size might not have changed and therefore the Resize won't trigger a layout.
	// Until we have a widget.Relayout() or similar, the renderer's refresh will do the re-layout.
	p.Refresh()
}

// Show this pop-up as overlay if not already shown.
func (p *PopUp) Show() {
	if !p.overlayShown {
		p.Canvas.Overlays().Add(p)
		p.overlayShown = true
	}
	p.Refresh()
	p.BaseWidget.Show()
}

// ShowAtPosition shows this pop-up at the given position.
func (p *PopUp) ShowAtPosition(pos fyne.Position) {
	p.Move(pos)
	p.Show()
}

// ShowAtRelativePosition shows this pop-up at the given position relative to stated object.
//
// Since 2.4
func (p *PopUp) ShowAtRelativePosition(rel fyne.Position, to fyne.CanvasObject) {
	withRelativePosition(rel, to, p.ShowAtPosition)
}

// Tapped is called when the user taps the popUp.
// If not modal and the tap is outside the content area, then dismiss this widget
func (p *PopUp) Tapped(e *fyne.PointEvent) {
	if !p.modal && !p.isInsideContent(e.Position) {
		p.Hide()
	}
}

// TappedSecondary is called when the user right/alt taps the popUp.
// If not modal and the tap is outside the content area, then dismiss this widget
func (p *PopUp) TappedSecondary(e *fyne.PointEvent) {
	if !p.modal && !p.isInsideContent(e.Position) {
		p.Hide()
	}
}

// MinSize returns the size that this widget should not shrink below
func (p *PopUp) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (p *PopUp) CreateRenderer() fyne.WidgetRenderer {
	th := p.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	p.ExtendBaseWidget(p)
	background := canvas.NewRectangle(th.Color(theme.ColorNameOverlayBackground, v))
	if p.modal {
		underlay := canvas.NewRectangle(th.Color(theme.ColorNameShadow, v))
		objects := []fyne.CanvasObject{underlay, background, p.Content}
		return &modalPopUpRenderer{
			widget.NewShadowingRenderer(objects, widget.DialogLevel),
			popUpBaseRenderer{popUp: p, background: background},
			underlay,
		}
	}
	objects := []fyne.CanvasObject{background, p.Content}
	return &popUpRenderer{
		widget.NewShadowingRenderer(objects, widget.PopUpLevel),
		popUpBaseRenderer{popUp: p, background: background},
	}
}

func (p *PopUp) isInsideContent(pos fyne.Position) bool {
	return pos.X >= p.innerPos.X && pos.Y >= p.innerPos.Y &&
		pos.X <= p.innerPos.X+p.innerSize.Width &&
		pos.Y <= p.innerPos.Y+p.innerSize.Height
}

// ShowPopUpAtPosition creates a new popUp for the specified content at the specified absolute position.
// It will then display the popup on the passed canvas.
func ShowPopUpAtPosition(content fyne.CanvasObject, canvas fyne.Canvas, pos fyne.Position) {
	newPopUp(content, canvas).ShowAtPosition(pos)
}

// ShowPopUpAtRelativePosition shows a new popUp for the specified content at the given position relative to stated object.
// It will then display the popup on the passed canvas.
//
// Since 2.4
func ShowPopUpAtRelativePosition(content fyne.CanvasObject, canvas fyne.Canvas, rel fyne.Position, to fyne.CanvasObject) {
	withRelativePosition(rel, to, func(pos fyne.Position) {
		ShowPopUpAtPosition(content, canvas, pos)
	})
}

func newPopUp(content fyne.CanvasObject, canvas fyne.Canvas) *PopUp {
	ret := &PopUp{Content: content, Canvas: canvas, modal: false}
	ret.ExtendBaseWidget(ret)
	return ret
}

// NewPopUp creates a new popUp for the specified content and displays it on the passed canvas.
func NewPopUp(content fyne.CanvasObject, canvas fyne.Canvas) *PopUp {
	return newPopUp(content, canvas)
}

// ShowPopUp creates a new popUp for the specified content and displays it on the passed canvas.
func ShowPopUp(content fyne.CanvasObject, canvas fyne.Canvas) {
	newPopUp(content, canvas).Show()
}

func newModalPopUp(content fyne.CanvasObject, canvas fyne.Canvas) *PopUp {
	p := &PopUp{Content: content, Canvas: canvas, modal: true}
	p.ExtendBaseWidget(p)
	return p
}

// NewModalPopUp creates a new popUp for the specified content and displays it on the passed canvas.
// A modal PopUp blocks interactions with underlying elements, covered with a semi-transparent overlay.
func NewModalPopUp(content fyne.CanvasObject, canvas fyne.Canvas) *PopUp {
	return newModalPopUp(content, canvas)
}

// ShowModalPopUp creates a new popUp for the specified content and displays it on the passed canvas.
// A modal PopUp blocks interactions with underlying elements, covered with a semi-transparent overlay.
func ShowModalPopUp(content fyne.CanvasObject, canvas fyne.Canvas) {
	p := newModalPopUp(content, canvas)
	p.Show()
}

type popUpBaseRenderer struct {
	popUp      *PopUp
	background *canvas.Rectangle
}

func (r *popUpBaseRenderer) padding() fyne.Size {
	th := r.popUp.Theme()
	return fyne.NewSquareSize(th.Size(theme.SizeNameInnerPadding))
}

func (r *popUpBaseRenderer) offset() fyne.Position {
	th := r.popUp.Theme()
	return fyne.NewSquareOffsetPos(th.Size(theme.SizeNameInnerPadding) / 2)
}

type popUpRenderer struct {
	*widget.ShadowingRenderer
	popUpBaseRenderer
}

func (r *popUpRenderer) Layout(_ fyne.Size) {
	innerSize := r.popUp.innerSize.Max(r.popUp.MinSize())
	r.popUp.Content.Resize(innerSize.Subtract(r.padding()))

	innerPos := r.popUp.innerPos
	if innerPos.X+innerSize.Width > r.popUp.Canvas.Size().Width {
		innerPos.X = r.popUp.Canvas.Size().Width - innerSize.Width
		if innerPos.X < 0 {
			innerPos.X = 0 // TODO here we may need a scroller as it's wider than our canvas
		}
	}
	if innerPos.Y+innerSize.Height > r.popUp.Canvas.Size().Height {
		innerPos.Y = r.popUp.Canvas.Size().Height - innerSize.Height
		if innerPos.Y < 0 {
			innerPos.Y = 0 // TODO here we may need a scroller as it's longer than our canvas
		}
	}
	r.popUp.Content.Move(innerPos.Add(r.offset()))

	r.background.Resize(innerSize)
	r.background.Move(innerPos)
	r.LayoutShadow(innerSize, innerPos)
}

func (r *popUpRenderer) MinSize() fyne.Size {
	return r.popUp.Content.MinSize().Add(r.padding())
}

func (r *popUpRenderer) Refresh() {
	th := r.popUp.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	r.background.FillColor = th.Color(theme.ColorNameOverlayBackground, v)
	expectedContentSize := r.popUp.innerSize.Max(r.popUp.MinSize()).Subtract(r.padding())
	shouldRelayout := r.popUp.Content.Size() != expectedContentSize

	if r.background.Size() != r.popUp.innerSize || r.background.Position() != r.popUp.innerPos || shouldRelayout {
		r.Layout(r.popUp.Size())
	}
	if r.popUp.Canvas.Size() != r.popUp.BaseWidget.Size() {
		r.popUp.BaseWidget.Resize(r.popUp.Canvas.Size())
	}
	r.popUp.Content.Refresh()
	r.background.Refresh()
	r.ShadowingRenderer.RefreshShadow()
}

type modalPopUpRenderer struct {
	*widget.ShadowingRenderer
	popUpBaseRenderer
	underlay *canvas.Rectangle
}

func (r *modalPopUpRenderer) Layout(canvasSize fyne.Size) {
	r.underlay.Resize(canvasSize)

	padding := r.padding()
	innerSize := r.popUp.innerSize.Max(r.popUp.Content.MinSize().Add(padding))

	requestedSize := innerSize.Subtract(padding)
	size := r.popUp.Content.MinSize().Max(requestedSize)
	size = size.Min(canvasSize.Subtract(padding))
	pos := fyne.NewPos((canvasSize.Width-size.Width)/2, (canvasSize.Height-size.Height)/2)
	r.popUp.Content.Move(pos)
	r.popUp.Content.Resize(size)

	innerPos := pos.Subtract(r.offset())
	r.background.Move(innerPos)
	r.background.Resize(size.Add(padding))
	r.LayoutShadow(innerSize, innerPos)
}

func (r *modalPopUpRenderer) MinSize() fyne.Size {
	return r.popUp.Content.MinSize().Add(r.padding())
}

func (r *modalPopUpRenderer) Refresh() {
	th := r.popUp.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	r.underlay.FillColor = th.Color(theme.ColorNameShadow, v)
	r.background.FillColor = th.Color(theme.ColorNameOverlayBackground, v)
	expectedContentSize := r.popUp.innerSize.Max(r.popUp.MinSize()).Subtract(r.padding())
	shouldLayout := r.popUp.Content.Size() != expectedContentSize

	if r.background.Size() != r.popUp.innerSize || shouldLayout {
		r.Layout(r.popUp.Size())
	}
	if r.popUp.Canvas.Size() != r.popUp.BaseWidget.Size() {
		r.popUp.BaseWidget.Resize(r.popUp.Canvas.Size())
	}
	r.popUp.Content.Refresh()
	r.background.Refresh()
}

func withRelativePosition(rel fyne.Position, to fyne.CanvasObject, f func(position fyne.Position)) {
	d := fyne.CurrentApp().Driver()
	c := d.CanvasForObject(to)
	if c == nil {
		fyne.LogError("Could not locate parent object to display relative to", nil)
		f(rel)
		return
	}

	pos := d.AbsolutePositionForObject(to).Add(rel)
	f(pos)
}
