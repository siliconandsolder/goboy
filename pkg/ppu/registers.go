package ppu

const (
	H_BLANK = iota
	V_BLANK
	OAM_SEARCH
	PIXEL_TRANSFER
)

type LcdControl struct {
	enabled         byte
	wTileMapArea    byte
	windowEnabled   byte
	tileDataArea    byte
	bgTileMapArea   byte
	objSize         byte
	objEnabled      byte
	bgWindowEnabled byte
}

func NewLcdControl() *LcdControl {
	return &LcdControl{
		enabled:         1,
		wTileMapArea:    0,
		windowEnabled:   0,
		tileDataArea:    1,
		bgTileMapArea:   0,
		objSize:         0,
		objEnabled:      0,
		bgWindowEnabled: 1,
	}
}

type LcdStatus struct {
	lycStatInterrupt    byte
	oamStatInterrupt    byte
	vBlankStatInterrupt byte
	hBlankStatInterrupt byte
	lycLYEqual          byte
	mode                byte
}

func NewLcdStatus() *LcdStatus {
	return &LcdStatus{
		lycStatInterrupt:    0,
		oamStatInterrupt:    0,
		vBlankStatInterrupt: 0,
		hBlankStatInterrupt: 0,
		lycLYEqual:          0,
		mode:                OAM_SEARCH,
	}
}

type ScrollStatus struct {
	scx byte
	scy byte
	wx  byte
	wy  byte
}

func NewScrollStatus() *ScrollStatus {
	return &ScrollStatus{
		scx: 0,
		scy: 0,
		wx:  0,
		wy:  0,
	}
}

type OamAttributes struct {
	priority byte
	yFlip    byte
	xFlip    byte
	palette  byte
}

type OamObj struct {
	posX       byte
	posY       byte
	tileNum    byte
	attributes OamAttributes
}

func NewOamObj() *OamObj {
	return &OamObj{
		posX:    0,
		posY:    0,
		tileNum: 0,
		attributes: OamAttributes{
			priority: 0,
			yFlip:    0,
			xFlip:    0,
			palette:  0,
		},
	}
}
