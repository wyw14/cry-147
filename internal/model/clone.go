package model

func CloneBytes(input []byte) []byte {
	if input == nil {
		return nil
	}
	out := make([]byte, len(input))
	copy(out, input)
	return out
}

func (cell *Cell) Clone() *Cell {
	if cell == nil {
		return nil
	}
	copyCell := *cell
	return &copyCell
}

func (tray *Tray) Clone() *Tray {
	if tray == nil {
		return nil
	}
	copyTray := *tray
	copyTray.Cells = CloneCells(tray.Cells)
	return &copyTray
}

func (sample *Sample) Clone() *Sample {
	if sample == nil {
		return nil
	}
	copySample := *sample
	copySample.Payload = CloneBytes(sample.Payload)
	return &copySample
}

func CloneSamples(samples []*Sample) []*Sample {
	out := make([]*Sample, 0, len(samples))
	for _, sample := range samples {
		out = append(out, sample.Clone())
	}
	return out
}

func CloneCells(cells map[string]*Cell) map[string]*Cell {
	out := make(map[string]*Cell, len(cells))
	for id, cell := range cells {
		out[id] = cell.Clone()
	}
	return out
}

func CloneCounts(counts map[string]int) map[string]int {
	out := make(map[string]int, len(counts))
	for key, value := range counts {
		out[key] = value
	}
	return out
}

func (result GradeResult) Clone() GradeResult {
	result.Counts = CloneCounts(result.Counts)
	return result
}
