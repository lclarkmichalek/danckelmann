package resources

import (
	"log"

	"github.com/bluepeppers/allegro"
)

const (
	DEFAULT_TILE_NAME   = "____DEFAULT____"
	DEFAULT_TILE_WIDTH  = 128
	DEFAULT_TILE_HEIGHT = 128
)

// Our custom super special bitmap class
type Bitmap struct {
	Raw *allegro.Bitmap
	// The offset of the bitmap when drawing
	OffX, OffY int
}

// Infomation about a tile in the atlas
type tileMetadata struct {
	index      int
	x, y, w, h int
	offx, offy int
	name       string
}

type ResourceManager struct {
	tileAtlas     *allegro.Bitmap
	tileMetadatas []tileMetadata
	tilePositions map[string]int
	tileSubs      map[string]*allegro.Bitmap

	fontMap map[string]*allegro.Font
}

func CreateResourceManager(config *ResourceManagerConfig) *ResourceManager {
	defaultTileConfig := TileConfig{Name: DEFAULT_TILE_NAME, Filename: DEFAULT_TILE_NAME}
	config.TileConfigs = append(config.TileConfigs, defaultTileConfig)

	tileConfs := make([]TileConfig, 0)
	for _, v := range config.TileConfigs {
		tileConfs = append(tileConfs, v)
	}
	tileBmps := make([]*allegro.Bitmap, len(tileConfs))
	tileMetadatas := make([]tileMetadata, len(tileConfs))
	maxHeight := 0
	totalWidth := 0
	for i := 0; i < len(tileBmps); i++ {
		cfg := tileConfs[i]

		// Load the bitmap
		var bmp *allegro.Bitmap
		if cfg.Name == DEFAULT_TILE_NAME {
			bmp = allegro.NewBitmap(DEFAULT_TILE_WIDTH, DEFAULT_TILE_HEIGHT)
		} else {
			bmp = allegro.LoadBitmap(cfg.Filename)
		}
		tileBmps[i] = bmp

		// Load the metadata, and then sanitize it
		x := cfg.X
		y := cfg.Y
		w := cfg.W
		h := cfg.H
		ox := cfg.OffX
		oy := cfg.OffY
		bmpw, bmph := bmp.GetDimensions()
		if bmpw < x {
			x = 0
			w = bmpw
		} else if bmpw < x+cfg.W {
			w = bmpw - x
		} else {
			w = cfg.W
		}
		if bmph < y {
			y = 0
			h = bmph
		} else if bmph < y+cfg.H {
			h = bmph - y
		} else {
			h = cfg.H
		}

		if w == 0 {
			w = bmpw - x
		}
		if h == 0 {
			h = bmph - y
		}

		if ox > w {
			ox = 0
		}
		if oy > h {
			oy = 0
		}

		if h > maxHeight {
			maxHeight = h
		}
		totalWidth += w

		tileMetadatas[i] = tileMetadata{i, x, y, w, h, ox, oy, cfg.Name}
	}

	// Draw the bitmaps to the atlas
	atlas := allegro.NewBitmap(totalWidth, maxHeight)
	atlas.SetTargetBitmap()
	allegro.HoldBitmapDrawing(true)

	currentPos := 0
	for i := 0; i < len(tileBmps); i++ {
		bmp := tileBmps[i]
		metadata := tileMetadatas[i]
		bmp.DrawRegion(float32(metadata.x), float32(metadata.y),
			float32(metadata.w), float32(metadata.h),
			float32(currentPos), float32(0), 0)
		
		// Update the metadata to the position in the atlas, from the position in the bmp
		metadata.x = currentPos
		metadata.y = 0
		currentPos += metadata.w
		tileMetadatas[i] = metadata
	}

	allegro.HoldBitmapDrawing(false)

	var manager ResourceManager
	manager.tileAtlas = atlas
	manager.tileMetadatas = tileMetadatas
	manager.tilePositions = make(map[string]int, len(tileMetadatas))
	manager.tileSubs = make(map[string]*allegro.Bitmap, len(tileMetadatas))
	// Populate the bmp cache
	for _, v := range tileMetadatas {
		manager.tilePositions[v.name] = v.index
		subBmp := manager.tileAtlas.CreateSubBitmap(v.x, v.y, v.w, v.h)
		manager.tileSubs[v.name] = subBmp
	}

	// Load the fonts
	manager.fontMap = make(map[string]*allegro.Font)
	for _, v := range config.FontConfigs {
		var font *allegro.Font
		if v.Filename == "builtin" {
			font = allegro.CreateBuiltinFont()
		} else {
			font = allegro.LoadFont(v.Filename, v.Size, 0)
		}
		manager.fontMap[v.Name] = font
	}

	return &manager
}

func (rm *ResourceManager) GetTile(name string) (*Bitmap, bool) {
	pos, ok := rm.tilePositions[name]
	if !ok {
		return nil, false
	}
	metadata := rm.tileMetadatas[pos]
	sub, ok := rm.tileSubs[name]
	if !ok {
		sub = rm.tileAtlas.CreateSubBitmap(metadata.x, metadata.y,
			metadata.w, metadata.h)
	}
	if sub == nil {
		return nil, false
	}
	bmp := &Bitmap{sub, metadata.offx, metadata.offy}
	
	return bmp, true
}

// Gets a tile that can be drawn, no matter what. Won't be pretty, but won't crash.
func (rm *ResourceManager) GetDefaultTile() *Bitmap {
	sub, ok := rm.GetTile(DEFAULT_TILE_NAME)
	if !ok {
		log.Panicf("Could not find default key in atlas: %v", DEFAULT_TILE_NAME)
	}
	return sub
}

func (rm *ResourceManager) GetTileOrDefault(name string) *Bitmap {
	tile, ok := rm.GetTile(name)
	if !ok {
		log.Printf("Could not find tile named %v. Defaulting to default tile.", name)
		return rm.GetDefaultTile()
	}
	return tile
}

func (rm *ResourceManager) GetFont(name string) (*allegro.Font, bool) {
	font, ok := rm.fontMap[name]
	return font, ok && font != nil
}
