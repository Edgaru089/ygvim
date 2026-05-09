package ui

import "github.com/Zyko0/go-sdl3/sdl"

// updateVertices re-generates vertices & indices for the grid.
func (grid *Grid) updateVertices(ui *UI) {

	grid.vertices = grid.vertices[:0]
	grid.indices = grid.indices[:0]

	for row, line := range grid.cells {
		for col, cell := range line {
			// first vertex indice
			id := int32(len(grid.vertices))
			offx, offy := ui.cellsize[0]*col, ui.cellsize[1]*row

			hl := ui.hl[cell.hlid]
			if hl.fg.A != 255 {
				hl = ui.hl[0]
			}
			if hl.bg.A != 255 {
				hl = ui.hl[0]
			}

			bg := sdl.FColor{
				R: float32(hl.fg.R) / 255.0,
				G: float32(hl.fg.G) / 255.0,
				B: float32(hl.fg.B) / 255.0,
				A: 1.0,
			}
			grid.vertices = append(grid.vertices,
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx), Y: float32(offy)},
					Color:    bg,
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx + ui.cellsize[0]), Y: float32(offy)},
					Color:    bg,
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx), Y: float32(offy + ui.cellsize[1])},
					Color:    bg,
				},
				sdl.Vertex{
					Position: sdl.FPoint{X: float32(offx + ui.cellsize[0]), Y: float32(offy + ui.cellsize[1])},
					Color:    bg,
				},
			)

			grid.indices = append(grid.indices, id, id+1, id+2, id+1, id+2, id+3)
		}
	}

}

func (ui *UI) Render(renderer *sdl.Renderer) {
	ui.Lock()
	defer ui.Unlock()

	for _, grid := range ui.grids {
		renderer.RenderGeometry(nil, grid.vertices, grid.indices)

		for row, line := range grid.cells {
			for col, cell := range line {
				offx, offy := ui.cellsize[0]*col, ui.cellsize[1]*row

				hl := ui.hl[cell.hlid]
				if hl.fg.A != 255 {
					hl = ui.hl[0]
				}
				if hl.bg.A != 255 {
					hl = ui.hl[0]
				}

				renderer.SetDrawColor(hl.fg.R, hl.fg.G, hl.fg.B, 255)
				renderer.SetDrawColor(0, 0, 0, 255)
				renderer.DebugText(float32(offx), float32(offy), cell.text)
			}
		}
	}
}
