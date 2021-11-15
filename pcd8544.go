package main

import (
	"github.com/stianeikeland/go-rpio/v4"

	"log"
)

/*
=====================================================
 Name: PCD8544.go
 Version: 0.2
 Copyright (C) 2021 sndnvaps<sndnvaps@gmail.com>

  Description :
     A simple PCD8544 LCD (Nokia3310/5110) for Raspberry Pi for displaying some system informations.
         Makes use of go-rpio of  Stian Eikeland (https://github.com/stianeikeland/go-rpio)

         Recommended connection (http://www.raspberrypi.org/archives/384):
         LCD pins          Raspberry Pi  GPIO
         LCD1 - GND        P06  -        GND
         LCD2 - VCC        P01 -         3.3V
         LCD3 - CLK        P11 -         GPIO17
         LCD4 - Din        P12 -         GPIO18
         LCD5 - D/C        P13 -         GPIO27
         LCD6 - CS         P15 -         GPIO22
         LCD7 - RST        P16 -         GPIO23
         LCD8 - LED        P01 -         3.3V
		 LCD9 - BlackLight P07 -         GPIO4
======================================================
*/
var (
	BLACK bool = true
	WHITE bool = false

	PCD8544_POWERDOWN           uint8 = 0x04
	PCD8544_ENTRYMODE           uint8 = 0x02
	PCD8544_EXTENDEDINSTRUCTION uint8 = 0x01

	PCD8544_DISPLAYBLANK    uint8 = 0x0
	PCD8544_DISPLAYNORMAL   uint8 = 0x4
	PCD8544_DISPLAYALLON    uint8 = 0x1
	PCD8544_DISPLAYINVERTED uint8 = 0x5

	// H = 0
	PCD8544_FUNCTIONSET    uint8 = 0x20
	PCD8544_DISPLAYCONTROL uint8 = 0x08
	PCD8544_SETYADDR       uint8 = 0x40
	PCD8544_SETXADDR       uint8 = 0x80

	// H = 1
	PCD8544_SETTEMP uint8 = 0x04
	PCD8544_SETBIAS uint8 = 0x10
	PCD8544_SETVOP  uint8 = 0x80

	// calibrate clock constants
	CLKCONST_1 = 8000
	CLKCONST_2 = 400 // 400 is a good tested value for Raspberry Pi

	// keywords
	LSBFIRST uint8 = 0
	MSBFIRST uint8 = 1

	textcolor bool
	cursor_x  uint8
	cursor_y  uint8
	textsize  uint8
)

type MonthType int

var (
	Jan MonthType = 1
	Feb MonthType = 2
	Mar MonthType = 3
	Apr MonthType = 4
	May MonthType = 5
	Jun MonthType = 6
	Jul MonthType = 7
	Aug MonthType = 8
	Sep MonthType = 9
	Oct MonthType = 10
	Nov MonthType = 11
	Dec MonthType = 12
)

func Month2String(m MonthType) string {
	switch m {
	case Jan:
		return "Jan"
	case Feb:
		return "Feb"
	case Mar:
		return "Mar"
	case Apr:
		return "Apr"
	case May:
		return "May"
	case Jun:
		return "Jun"
	case Jul:
		return "Jul"
	case Aug:
		return "Aug"
	case Sep:
		return "Sep"
	case Oct:
		return "Oct"
	case Nov:
		return "Nov"
	case Dec:
		return "Dec"
	}
	return "Non"
}

func init() {
	//open gpio
	err := rpio.Open()
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range FONTS {
		dict.Set(byte(key)+0x20, value)
	}
	dict.Set(byte(0xb0), [5]byte{0x00, 0x06, 0x06, 0x00, 0x00}) // 0xb0 °

}

type PCD8544_pin struct {
	_din  uint8
	_sclk uint8
	_dc   uint8
	_rst  uint8
	_cs   uint8
	_bl   uint8
}

var dict *ByteDictionary = NewByteDictionary()

/** @array Charset */
var FONTS [][5]byte = [][5]byte{
	{0x00, 0x00, 0x00, 0x00, 0x00}, // 20 space
	{0x81, 0x81, 0x18, 0x81, 0x81}, // 21 !
	{0x00, 0x07, 0x00, 0x07, 0x00}, // 22 "
	{0x14, 0x7f, 0x14, 0x7f, 0x14}, // 23 #
	{0x24, 0x2a, 0x7f, 0x2a, 0x12}, // 24 $
	{0x23, 0x13, 0x08, 0x64, 0x62}, // 25 %
	{0x36, 0x49, 0x55, 0x22, 0x50}, // 26 &
	{0x00, 0x05, 0x03, 0x00, 0x00}, // 27 '
	{0x00, 0x1c, 0x22, 0x41, 0x00}, // 28 (
	{0x00, 0x41, 0x22, 0x1c, 0x00}, // 29 )
	{0x14, 0x08, 0x3e, 0x08, 0x14}, // 2a *
	{0x08, 0x08, 0x3e, 0x08, 0x08}, // 2b +
	{0x00, 0x50, 0x30, 0x00, 0x00}, // 2c ,
	{0x08, 0x08, 0x08, 0x08, 0x08}, // 2d -
	{0x00, 0x60, 0x60, 0x00, 0x00}, // 2e .
	{0x20, 0x10, 0x08, 0x04, 0x02}, // 2f /
	{0x3e, 0x51, 0x49, 0x45, 0x3e}, // 30 0
	{0x00, 0x42, 0x7f, 0x40, 0x00}, // 31 1
	{0x42, 0x61, 0x51, 0x49, 0x46}, // 32 2
	{0x21, 0x41, 0x45, 0x4b, 0x31}, // 33 3
	{0x18, 0x14, 0x12, 0x7f, 0x10}, // 34 4
	{0x27, 0x45, 0x45, 0x45, 0x39}, // 35 5
	{0x3c, 0x4a, 0x49, 0x49, 0x30}, // 36 6
	{0x01, 0x71, 0x09, 0x05, 0x03}, // 37 7
	{0x36, 0x49, 0x49, 0x49, 0x36}, // 38 8
	{0x06, 0x49, 0x49, 0x29, 0x1e}, // 39 9
	{0x00, 0x36, 0x36, 0x00, 0x00}, // 3a :
	{0x00, 0x56, 0x36, 0x00, 0x00}, // 3b ;
	{0x08, 0x14, 0x22, 0x41, 0x00}, // 3c <
	{0x14, 0x14, 0x14, 0x14, 0x14}, // 3d =
	{0x00, 0x41, 0x22, 0x14, 0x08}, // 3e >
	{0x02, 0x01, 0x51, 0x09, 0x06}, // 3f ?
	{0x32, 0x49, 0x79, 0x41, 0x3e}, // 40 @
	{0x7e, 0x11, 0x11, 0x11, 0x7e}, // 41 A
	{0x7f, 0x49, 0x49, 0x49, 0x36}, // 42 B
	{0x3e, 0x41, 0x41, 0x41, 0x22}, // 43 C
	{0x7f, 0x41, 0x41, 0x22, 0x1c}, // 44 D
	{0x7f, 0x49, 0x49, 0x49, 0x41}, // 45 E
	{0x7f, 0x09, 0x09, 0x09, 0x01}, // 46 F
	{0x3e, 0x41, 0x49, 0x49, 0x7a}, // 47 G
	{0x7f, 0x08, 0x08, 0x08, 0x7f}, // 48 H
	{0x00, 0x41, 0x7f, 0x41, 0x00}, // 49 I
	{0x20, 0x40, 0x41, 0x3f, 0x01}, // 4a J
	{0x7f, 0x08, 0x14, 0x22, 0x41}, // 4b K
	{0x7f, 0x40, 0x40, 0x40, 0x40}, // 4c L
	{0x7f, 0x02, 0x0c, 0x02, 0x7f}, // 4d M
	{0x7f, 0x04, 0x08, 0x10, 0x7f}, // 4e N
	{0x3e, 0x41, 0x41, 0x41, 0x3e}, // 4f O
	{0x7f, 0x09, 0x09, 0x09, 0x06}, // 50 P
	{0x3e, 0x41, 0x51, 0x21, 0x5e}, // 51 Q
	{0x7f, 0x09, 0x19, 0x29, 0x46}, // 52 R
	{0x46, 0x49, 0x49, 0x49, 0x31}, // 53 S
	{0x01, 0x01, 0x7f, 0x01, 0x01}, // 54 T
	{0x3f, 0x40, 0x40, 0x40, 0x3f}, // 55 U
	{0x1f, 0x20, 0x40, 0x20, 0x1f}, // 56 V
	{0x3f, 0x40, 0x38, 0x40, 0x3f}, // 57 W
	{0x63, 0x14, 0x08, 0x14, 0x63}, // 58 X
	{0x07, 0x08, 0x70, 0x08, 0x07}, // 59 Y
	{0x61, 0x51, 0x49, 0x45, 0x43}, // 5a Z
	{0x00, 0x7f, 0x41, 0x41, 0x00}, // 5b [
	{0x02, 0x04, 0x08, 0x10, 0x20}, // 5c backslash
	{0x00, 0x41, 0x41, 0x7f, 0x00}, // 5d ]
	{0x04, 0x02, 0x01, 0x02, 0x04}, // 5e ^
	{0x40, 0x40, 0x40, 0x40, 0x40}, // 5f _
	{0x00, 0x01, 0x02, 0x04, 0x00}, // 60 `
	{0x20, 0x54, 0x54, 0x54, 0x78}, // 61 a
	{0x7f, 0x48, 0x44, 0x44, 0x38}, // 62 b
	{0x38, 0x44, 0x44, 0x44, 0x20}, // 63 c
	{0x38, 0x44, 0x44, 0x48, 0x7f}, // 64 d
	{0x38, 0x54, 0x54, 0x54, 0x18}, // 65 e
	{0x08, 0x7e, 0x09, 0x01, 0x02}, // 66 f
	{0x0c, 0x52, 0x52, 0x52, 0x3e}, // 67 g
	{0x7f, 0x08, 0x04, 0x04, 0x78}, // 68 h
	{0x00, 0x44, 0x7d, 0x40, 0x00}, // 69 i
	{0x20, 0x40, 0x44, 0x3d, 0x00}, // 6a j
	{0x7f, 0x10, 0x28, 0x44, 0x00}, // 6b k
	{0x00, 0x41, 0x7f, 0x40, 0x00}, // 6c l
	{0x7c, 0x04, 0x18, 0x04, 0x78}, // 6d m
	{0x7c, 0x08, 0x04, 0x04, 0x78}, // 6e n
	{0x38, 0x44, 0x44, 0x44, 0x38}, // 6f o
	{0x7c, 0x14, 0x14, 0x14, 0x08}, // 70 p
	{0x08, 0x14, 0x14, 0x14, 0x7c}, // 71 q
	{0x7c, 0x08, 0x04, 0x04, 0x08}, // 72 r
	{0x48, 0x54, 0x54, 0x54, 0x20}, // 73 s
	{0x04, 0x3f, 0x44, 0x40, 0x20}, // 74 t
	{0x3c, 0x40, 0x40, 0x20, 0x7c}, // 75 u
	{0x1c, 0x20, 0x40, 0x20, 0x1c}, // 76 v
	{0x3c, 0x40, 0x30, 0x40, 0x3c}, // 77 w
	{0x44, 0x28, 0x10, 0x28, 0x44}, // 78 x
	{0x0c, 0x50, 0x50, 0x50, 0x3c}, // 79 y
	{0x44, 0x64, 0x54, 0x4c, 0x44}, // 7a z
	{0x00, 0x08, 0x36, 0x41, 0x00}, // 7b {
	{0x00, 0x00, 0x7f, 0x00, 0x00}, // 7c |
	{0x00, 0x41, 0x36, 0x08, 0x00}, // 7d }
	{0x10, 0x08, 0x08, 0x10, 0x08}, // 7e ~
	{0x00, 0x00, 0x00, 0x00, 0x00}, // 7f
	//DB 00H,40H,00H,00H,00H,00H;"°",0 //b0 °
	//{0x00,0x40,0x00,0x00,0x00}, //b0 °
}

const LCDWIDTH uint8 = 84
const LCDHEIGHT uint8 = 48

// the memory buffer for the LCD
// pcd8544 have 6 Page * 84 Column
//
// +-------------------------------------+ <-
// |               page 0                |  |
// +-------------------------------------+  |
// |               page 1                |  |
// +-------------------------------------+  |
// |               page 2                |  |
// +-------------------------------------+  | 8*6=48 Dot
// |               page 3                |  |
// +-------------------------------------+  |
// |               page 4                |  |
// +-------------------------------------+  |
// |               page 5                |  |
// +-------------------------------------+ <-
// ^                                     ^
// 0<---------------- 84 dot ----------->83
//
// each page have 84 byte
// +-------------------------------------+
// |         page n = 84 byte            |
// +-------------------------------------+
//

// the memory buffer for the LCD
var pcd8544_buffer [6][LCDWIDTH]byte

var pi_logo []byte = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0010 (16) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0xF8, 0xFC, 0xAE, 0x0E, 0x0E, 0x06, 0x0E, 0x06, // 0x0020 (32) pixels
	0xCE, 0x86, 0x8E, 0x0E, 0x0E, 0x1C, 0xB8, 0xF0, 0xF8, 0x78, 0x38, 0x1E, 0x0E, 0x8E, 0x8E, 0xC6, // 0x0030 (48) pixels
	0x0E, 0x06, 0x0E, 0x06, 0x0E, 0x9E, 0xFE, 0xFC, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0040 (64) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0050 (80) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0060 (96) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x0F, 0x0F, 0xFE, // 0x0070 (112) pixels
	0xF8, 0xF0, 0x60, 0x60, 0xE0, 0xE1, 0xE3, 0xF7, 0x7E, 0x3E, 0x1E, 0x1F, 0x1F, 0x1F, 0x3E, 0x7E, // 0x0080 (128) pixels
	0xFB, 0xF3, 0xE1, 0xE0, 0x60, 0x70, 0xF0, 0xF8, 0xBE, 0x1F, 0x0F, 0x07, 0x00, 0x00, 0x00, 0x00, // 0x0090 (144) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x00A0 (160) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x00B0 (176) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xC0, // 0x00C0 (192) pixels
	0xE0, 0xFC, 0xFE, 0xFF, 0xF3, 0x38, 0x38, 0x0C, 0x0E, 0x0F, 0x0F, 0x0F, 0x0E, 0x3C, 0x38, 0xF8, // 0x00D0 (208) pixels
	0xF8, 0x38, 0x3C, 0x0E, 0x0F, 0x0F, 0x0F, 0x0E, 0x0C, 0x38, 0x38, 0xF3, 0xFF, 0xFF, 0xF8, 0xE0, // 0x00E0 (224) pixels
	0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x00F0 (240) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0100 (256) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0110 (272) pixels
	0x00, 0x7F, 0xFF, 0xE7, 0xC3, 0xC1, 0xE0, 0xFF, 0xFF, 0x78, 0xE0, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, // 0x0120 (288) pixels
	0x60, 0x78, 0x38, 0x3F, 0x3F, 0x38, 0x38, 0x60, 0x60, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, 0xF8, 0x7F, // 0x0130 (304) pixels
	0xFF, 0xE0, 0xC1, 0xC3, 0xE7, 0x7F, 0x3E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0140 (320) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0150 (336) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0160 (352) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x0F, 0x7F, 0xFF, 0xF1, 0xE0, 0xC0, 0x80, 0x01, // 0x0170 (368) pixels
	0x03, 0x9F, 0xFF, 0xF0, 0xE0, 0xE0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, 0xE0, 0xF0, 0xFF, 0x9F, // 0x0180 (384) pixels
	0x03, 0x01, 0x80, 0xC0, 0xE0, 0xF1, 0x7F, 0x1F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x0190 (400) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01A0 (416) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01B0 (432) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // 0x01C0 (448) pixels
	0x03, 0x03, 0x07, 0x07, 0x0F, 0x1F, 0x1F, 0x3F, 0x3B, 0x71, 0x60, 0x60, 0x60, 0x60, 0x60, 0x71, // 0x01D0 (464) pixels
	0x3B, 0x1F, 0x0F, 0x0F, 0x0F, 0x07, 0x03, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01E0 (480) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 0x01F0 (496) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func LCDInit(SCLK, DIN, DC, CS, RST, BL, contrast uint8) (pin PCD8544_pin) {
	_din := DIN
	_sclk := SCLK
	_dc := DC
	_rst := RST
	_cs := CS
	_bl := BL
	_contrast := contrast

	pin = PCD8544_pin{
		_din:  _din,
		_sclk: _sclk,
		_dc:   _dc,
		_rst:  _rst,
		_cs:   _cs,
		_bl:   _bl,
	}

	cursor_x = 0
	cursor_y = 0
	textsize = 1
	textcolor = BLACK

	dinPin := rpio.Pin(_din)
	sclkPin := rpio.Pin(_sclk)
	dcPin := rpio.Pin(_dc)
	rstPin := rpio.Pin(_rst)
	csPin := rpio.Pin(_cs)
	blPin := rpio.Pin(_bl)

	//set output mode
	dinPin.Output()
	sclkPin.Output()
	dcPin.Output()
	rstPin.Output()
	csPin.Output()
	blPin.Output()

	//Light on
	blPin.High()

	// toggle RST low to reset; CS low so it'll listen to us
	if _cs > 0 {
		csPin.Low()
	}

	rstPin.Low()

	//sleep 500ms
	_delay_ms(500)
	//time.Sleep(time.Duration(500) * time.Millisecond)
	rstPin.High()

	// get into the EXTENDED mode!
	pin.LCDCommand(PCD8544_FUNCTIONSET | PCD8544_EXTENDEDINSTRUCTION)

	// LCD bias select (4 is optimal?)
	pin.LCDCommand(PCD8544_SETBIAS | 0x4)

	// set VOP
	if _contrast > 0x7f {
		_contrast = 0x7f
	}

	pin.LCDCommand(PCD8544_SETVOP | _contrast) // Experimentally determined

	// normal mode
	pin.LCDCommand(PCD8544_FUNCTIONSET)

	// Set display to Normal
	pin.LCDCommand(PCD8544_DISPLAYCONTROL | PCD8544_DISPLAYNORMAL)

	// set up a bounding box for screen updates
	//updateBoundingBox(0, 0, LCDWIDTH-1, LCDHEIGHT-1);

	return pin

}

func (pin PCD8544_pin) LCDCommand(cmd uint8) {
	//dcPin.Low()
	pin.LCDspiwrite(0, cmd)
}

//往LCD写入数据
//data_cmd: 1 -> 数据， 0 -> 命令
//val： 需要写入的数据
func (pin PCD8544_pin) LCDspiwrite(data_cmd uint8, val uint8) {
	csPin := rpio.Pin(pin._cs)
	dinPin := rpio.Pin(pin._din)
	sclkPin := rpio.Pin(pin._sclk)
	dcPin := rpio.Pin(pin._dc)
	csPin.Low()
	if data_cmd == 1 { //写入数据
		rpio.WritePin(dcPin, rpio.High)
	} else { //写入命令
		rpio.WritePin(dcPin, rpio.Low)
	}
	for i := 0; i < 8; i++ {
		if (val & 0x80) == 0 {
			rpio.WritePin(dinPin, rpio.Low)
		} else {
			rpio.WritePin(dinPin, rpio.High)
		}
		rpio.WritePin(sclkPin, rpio.Low)
		val = val << 1
		rpio.WritePin(sclkPin, rpio.High)
	}

	csPin.High()

}

func (pin PCD8544_pin) LCDData(c uint8) {
	pin.LCDspiwrite(1, c)
}

func (pin PCD8544_pin) LCDSetcontrast(val uint8) {
	if val > 0x7f {
		val = 0x7f
	}
	pin.LCDCommand(PCD8544_FUNCTIONSET | PCD8544_EXTENDEDINSTRUCTION)
	pin.LCDCommand(PCD8544_SETVOP | val)
	pin.LCDCommand(PCD8544_FUNCTIONSET)
}

func (pin PCD8544_pin) LCDDisplay() {
	var col, maxcol, p uint8
	for p = 0; p < 6; p++ {
		pin.LCDCommand(PCD8544_SETYADDR | p)
		// start at the beginning of the row
		col = 0
		maxcol = LCDWIDTH - 1
		pin.LCDCommand(PCD8544_SETXADDR | col)

		for ; col <= maxcol; col++ {
			//uart_putw_dec(col);
			//uart_putchar(' ');
			pin.LCDData(pcd8544_buffer[p][col])
		}
	}
	pin.LCDCommand(PCD8544_SETYADDR) // no idea why this is necessary but it is to finish the last byte?
}

func (pin PCD8544_pin) LCDShowRpiLogo() {
	var i int
	for i = 0; i < 6; i++ {
		/*
			pi_logo_slice -> pi_logo[0:84]
			                 pi_logo[84:168]
							 pi_logo[168:252]
							 pi_logo[252:336]
							 pi_logo[336:420]
							 pi_logo[420:504]
		*/
		pi_logo_slice := pi_logo[(i * (len(pi_logo) / 6)):((i + 1) * 84)]
		copy(pcd8544_buffer[i][:], pi_logo_slice[:])
	}
	pin.LCDDisplay()
}

func LCDClear() {
	var i uint8
	var j uint8

	for i = 0; i < 6; i++ {
		for j = 0; j < LCDWIDTH-1; j++ {
			pcd8544_buffer[i][j] = 0
		}
	}
	cursor_y = 0
	cursor_x = 0
}

func _delay_ms(t uint32) {
	nCount := CLKCONST_1

	for {
		if t == 0 {
			break
		}

		for {
			if nCount == 0 {
				break
			}
			nCount--
		}
		t--

	}
}

/*
 y = uint8{0,1,2,3,4,5}
*/
func LCDDrawString(x uint8, y uint8, val []byte) {
	cursor_x = x
	cursor_y = y
	//setup for debug
	//	fmt.Printf("LCDDrawString -> val = %s\n",string(val))
	for i := 0; i < len(val); i++ {
		LCDWrite(val[i])
	}
}

func LCDWrite(c byte) {

	if c == '\n' {
		cursor_y += textsize * 8
		cursor_x = 0
	} else if c == '\r' {
		//skip em
	} else {
		LCDDrawchar(cursor_x, cursor_y, c)
		cursor_x += textsize * 6
		if cursor_x >= (LCDWIDTH - 5) {
			cursor_x = 0
			cursor_y += 8
		}
		if cursor_y >= LCDHEIGHT {
			cursor_y = 0

		}
	}
}

func LCDDrawchar(x uint8, y uint8, c byte) int {
	if y >= LCDHEIGHT {
		return 0
	}
	if x+5 >= LCDWIDTH {
		return 0
	}
	if c < 0x20 || c > 0xb3 {
		// out of range
		return 0
	}
	var i uint8
	for i = 0; i < 5; i++ {
		charIndex := c
		//pcd8544_buffer[y][x+i] = FONTS[charIndex][i]
		pcd8544_buffer[y][x+i] = dict.Get(charIndex)[i]

	}
	return int(x + 6)

}

func LCDDrawPixel(x uint8, y uint8) {
	pcd8544_buffer[y>>3][x] |= 1 << (y % 8)

}

func Swap(x *uint8, y *uint8) {

	temp := *x
	*x = *y
	*y = temp
}

func Abs(n uint8) uint8 {
	if n < 0 {
		return (-n)
	}
	return n
}

func LCDDrawLine(x0 uint8, y0 uint8, x1 uint8, y1 uint8) {
	var (
		steep               bool
		deltax, deltay, err uint8
		x, y                uint8
		ystep               int8
	)

	// simple clipping is done in the drawPixel routine
	/*
	   if (( x0 < 0) || (x0 > HRES)) return;
	   if (( x1 < 0) || (x1 > HRES)) return;
	   if (( y0 < 0) || (y0 > VRES)) return;
	   if (( y1 < 0) || (y1 > VRES)) return;
	*/

	steep = Abs(y1-y0) > Abs(x1-x0)

	if steep {
		Swap(&x0, &y0)
		Swap(&x1, &y1)
	}

	if x0 > x1 {
		Swap(&x0, &x1)
		Swap(&y0, &y1)
	}

	deltax = x1 - x0
	deltay = Abs(y1 - y0)

	err = 0
	y = y0

	if y0 < y1 {
		ystep = 1
	} else {
		ystep = -1
	}

	for x = x0; x < x1; x++ {
		if steep {
			LCDDrawPixel(y, x)
		} else {
			LCDDrawPixel(x, y)
		}
		err += deltay

		if (err << 1) >= deltax {

			y += (uint8)(ystep)
			err -= deltax
		}
	}
}

func LCDDrawVLine(x uint8, y uint8, h uint8) {
	LCDDrawLine(x, y, x, y+h-1)
}

func LCDDrawHLine(x uint8, y uint8, w uint8) {
	LCDDrawLine(x, y, x+w-1, y)
}

func LCDDrawTriangle(x1 uint8, y1 uint8, x2 uint8, y2 uint8, x3 uint8, y3 uint8) {
	LCDDrawLine(x1, y1, x2, y2)
	LCDDrawLine(x2, y2, x3, y3)
	LCDDrawLine(x3, y3, x1, y1)
}
