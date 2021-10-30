package main

import (
	"github.com/stianeikeland/go-rpio"
	"time"
	"log"
	"fmt"
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

	PCD8544_POWERDOWN uint8 = 0x04
	PCD8544_ENTRYMODE uint8 = 0x02
	PCD8544_EXTENDEDINSTRUCTION uint8 = 0x01

	PCD8544_DISPLAYBLANK uint8 = 0x0
	PCD8544_DISPLAYNORMAL uint8 = 0x4
	PCD8544_DISPLAYALLON uint8 = 0x1
	PCD8544_DISPLAYINVERTED uint8  = 0x5
	
	// H = 0
	PCD8544_FUNCTIONSET uint8 = 0x20
	PCD8544_DISPLAYCONTROL uint8 = 0x08
	PCD8544_SETYADDR uint8 = 0x40
	PCD8544_SETXADDR uint8  = 0x80
	
	// H = 1
	PCD8544_SETTEMP uint8 =  0x04
	PCD8544_SETBIAS uint8 =  0x10
	PCD8544_SETVOP uint8 = 0x80

     // calibrate clock constants
    CLKCONST_1 = 8000
    CLKCONST_2 = 400  // 400 is a good tested value for Raspberry Pi

    // keywords
    LSBFIRST uint8 = 0
    MSBFIRST uint8 = 1

	textcolor bool
	cursor_x uint8
	cursor_y uint8
	textsize uint8

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
}

func swap(a,b *int) {
	t := *a
    *a = *b
	*b = t
}

func _BV(bit uint8) uint8 {
	return (0x1 << (bit))
}
type PCD8544_pin struct {
     _din uint8
	 _sclk uint8
	 _dc uint8
	 _rst uint8
	 _cs uint8
	 _bl uint8

}

// font bitmap
 var  font []byte  = []byte {
	0x00, 0x00, 0x00, 0x00, 0x00,
	0x3E, 0x5B, 0x4F, 0x5B, 0x3E,
	0x3E, 0x6B, 0x4F, 0x6B, 0x3E,
	0x1C, 0x3E, 0x7C, 0x3E, 0x1C,
	0x18, 0x3C, 0x7E, 0x3C, 0x18,
	0x1C, 0x57, 0x7D, 0x57, 0x1C,
	0x1C, 0x5E, 0x7F, 0x5E, 0x1C,
	0x00, 0x18, 0x3C, 0x18, 0x00,
	0xFF, 0xE7, 0xC3, 0xE7, 0xFF,
	0x00, 0x18, 0x24, 0x18, 0x00,
	0xFF, 0xE7, 0xDB, 0xE7, 0xFF,
	0x30, 0x48, 0x3A, 0x06, 0x0E,
	0x26, 0x29, 0x79, 0x29, 0x26,
	0x40, 0x7F, 0x05, 0x05, 0x07,
	0x40, 0x7F, 0x05, 0x25, 0x3F,
	0x5A, 0x3C, 0xE7, 0x3C, 0x5A,
	0x7F, 0x3E, 0x1C, 0x1C, 0x08,
	0x08, 0x1C, 0x1C, 0x3E, 0x7F,
	0x14, 0x22, 0x7F, 0x22, 0x14,
	0x5F, 0x5F, 0x00, 0x5F, 0x5F,
	0x06, 0x09, 0x7F, 0x01, 0x7F,
	0x00, 0x66, 0x89, 0x95, 0x6A,
	0x60, 0x60, 0x60, 0x60, 0x60,
	0x94, 0xA2, 0xFF, 0xA2, 0x94,
	0x08, 0x04, 0x7E, 0x04, 0x08,
	0x10, 0x20, 0x7E, 0x20, 0x10,
	0x08, 0x08, 0x2A, 0x1C, 0x08,
	0x08, 0x1C, 0x2A, 0x08, 0x08,
	0x1E, 0x10, 0x10, 0x10, 0x10,
	0x0C, 0x1E, 0x0C, 0x1E, 0x0C,
	0x30, 0x38, 0x3E, 0x38, 0x30,
	0x06, 0x0E, 0x3E, 0x0E, 0x06,
	0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x5F, 0x00, 0x00,
	0x00, 0x07, 0x00, 0x07, 0x00,
	0x14, 0x7F, 0x14, 0x7F, 0x14,
	0x24, 0x2A, 0x7F, 0x2A, 0x12,
	0x23, 0x13, 0x08, 0x64, 0x62,
	0x36, 0x49, 0x56, 0x20, 0x50,
	0x00, 0x08, 0x07, 0x03, 0x00,
	0x00, 0x1C, 0x22, 0x41, 0x00,
	0x00, 0x41, 0x22, 0x1C, 0x00,
	0x2A, 0x1C, 0x7F, 0x1C, 0x2A,
	0x08, 0x08, 0x3E, 0x08, 0x08,
	0x00, 0x80, 0x70, 0x30, 0x00,
	0x08, 0x08, 0x08, 0x08, 0x08,
	0x00, 0x00, 0x60, 0x60, 0x00,
	0x20, 0x10, 0x08, 0x04, 0x02,
	0x3E, 0x51, 0x49, 0x45, 0x3E,
	0x00, 0x42, 0x7F, 0x40, 0x00,
	0x72, 0x49, 0x49, 0x49, 0x46,
	0x21, 0x41, 0x49, 0x4D, 0x33,
	0x18, 0x14, 0x12, 0x7F, 0x10,
	0x27, 0x45, 0x45, 0x45, 0x39,
	0x3C, 0x4A, 0x49, 0x49, 0x31,
	0x41, 0x21, 0x11, 0x09, 0x07,
	0x36, 0x49, 0x49, 0x49, 0x36,
	0x46, 0x49, 0x49, 0x29, 0x1E,
	0x00, 0x00, 0x14, 0x00, 0x00,
	0x00, 0x40, 0x34, 0x00, 0x00,
	0x00, 0x08, 0x14, 0x22, 0x41,
	0x14, 0x14, 0x14, 0x14, 0x14,
	0x00, 0x41, 0x22, 0x14, 0x08,
	0x02, 0x01, 0x59, 0x09, 0x06,
	0x3E, 0x41, 0x5D, 0x59, 0x4E,
	0x7C, 0x12, 0x11, 0x12, 0x7C,
	0x7F, 0x49, 0x49, 0x49, 0x36,
	0x3E, 0x41, 0x41, 0x41, 0x22,
	0x7F, 0x41, 0x41, 0x41, 0x3E,
	0x7F, 0x49, 0x49, 0x49, 0x41,
	0x7F, 0x09, 0x09, 0x09, 0x01,
	0x3E, 0x41, 0x41, 0x51, 0x73,
	0x7F, 0x08, 0x08, 0x08, 0x7F,
	0x00, 0x41, 0x7F, 0x41, 0x00,
	0x20, 0x40, 0x41, 0x3F, 0x01,
	0x7F, 0x08, 0x14, 0x22, 0x41,
	0x7F, 0x40, 0x40, 0x40, 0x40,
	0x7F, 0x02, 0x1C, 0x02, 0x7F,
	0x7F, 0x04, 0x08, 0x10, 0x7F,
	0x3E, 0x41, 0x41, 0x41, 0x3E,
	0x7F, 0x09, 0x09, 0x09, 0x06,
	0x3E, 0x41, 0x51, 0x21, 0x5E,
	0x7F, 0x09, 0x19, 0x29, 0x46,
	0x26, 0x49, 0x49, 0x49, 0x32,
	0x03, 0x01, 0x7F, 0x01, 0x03,
	0x3F, 0x40, 0x40, 0x40, 0x3F,
	0x1F, 0x20, 0x40, 0x20, 0x1F,
	0x3F, 0x40, 0x38, 0x40, 0x3F,
	0x63, 0x14, 0x08, 0x14, 0x63,
	0x03, 0x04, 0x78, 0x04, 0x03,
	0x61, 0x59, 0x49, 0x4D, 0x43,
	0x00, 0x7F, 0x41, 0x41, 0x41,
	0x02, 0x04, 0x08, 0x10, 0x20,
	0x00, 0x41, 0x41, 0x41, 0x7F,
	0x04, 0x02, 0x01, 0x02, 0x04,
	0x40, 0x40, 0x40, 0x40, 0x40,
	0x00, 0x03, 0x07, 0x08, 0x00,
	0x20, 0x54, 0x54, 0x78, 0x40,
	0x7F, 0x28, 0x44, 0x44, 0x38,
	0x38, 0x44, 0x44, 0x44, 0x28,
	0x38, 0x44, 0x44, 0x28, 0x7F,
	0x38, 0x54, 0x54, 0x54, 0x18,
	0x00, 0x08, 0x7E, 0x09, 0x02,
	0x18, 0xA4, 0xA4, 0x9C, 0x78,
	0x7F, 0x08, 0x04, 0x04, 0x78,
	0x00, 0x44, 0x7D, 0x40, 0x00,
	0x20, 0x40, 0x40, 0x3D, 0x00,
	0x7F, 0x10, 0x28, 0x44, 0x00,
	0x00, 0x41, 0x7F, 0x40, 0x00,
	0x7C, 0x04, 0x78, 0x04, 0x78,
	0x7C, 0x08, 0x04, 0x04, 0x78,
	0x38, 0x44, 0x44, 0x44, 0x38,
	0xFC, 0x18, 0x24, 0x24, 0x18,
	0x18, 0x24, 0x24, 0x18, 0xFC,
	0x7C, 0x08, 0x04, 0x04, 0x08,
	0x48, 0x54, 0x54, 0x54, 0x24,
	0x04, 0x04, 0x3F, 0x44, 0x24,
	0x3C, 0x40, 0x40, 0x20, 0x7C,
	0x1C, 0x20, 0x40, 0x20, 0x1C,
	0x3C, 0x40, 0x30, 0x40, 0x3C,
	0x44, 0x28, 0x10, 0x28, 0x44,
	0x4C, 0x90, 0x90, 0x90, 0x7C,
	0x44, 0x64, 0x54, 0x4C, 0x44,
	0x00, 0x08, 0x36, 0x41, 0x00,
	0x00, 0x00, 0x77, 0x00, 0x00,
	0x00, 0x41, 0x36, 0x08, 0x00,
	0x02, 0x01, 0x02, 0x04, 0x02,
	0x3C, 0x26, 0x23, 0x26, 0x3C,
	0x1E, 0xA1, 0xA1, 0x61, 0x12,
	0x3A, 0x40, 0x40, 0x20, 0x7A,
	0x38, 0x54, 0x54, 0x55, 0x59,
	0x21, 0x55, 0x55, 0x79, 0x41,
	0x21, 0x54, 0x54, 0x78, 0x41,
	0x21, 0x55, 0x54, 0x78, 0x40,
	0x20, 0x54, 0x55, 0x79, 0x40,
	0x0C, 0x1E, 0x52, 0x72, 0x12,
	0x39, 0x55, 0x55, 0x55, 0x59,
	0x39, 0x54, 0x54, 0x54, 0x59,
	0x39, 0x55, 0x54, 0x54, 0x58,
	0x00, 0x00, 0x45, 0x7C, 0x41,
	0x00, 0x02, 0x45, 0x7D, 0x42,
	0x00, 0x01, 0x45, 0x7C, 0x40,
	0xF0, 0x29, 0x24, 0x29, 0xF0,
	0xF0, 0x28, 0x25, 0x28, 0xF0,
	0x7C, 0x54, 0x55, 0x45, 0x00,
	0x20, 0x54, 0x54, 0x7C, 0x54,
	0x7C, 0x0A, 0x09, 0x7F, 0x49,
	0x32, 0x49, 0x49, 0x49, 0x32,
	0x32, 0x48, 0x48, 0x48, 0x32,
	0x32, 0x4A, 0x48, 0x48, 0x30,
	0x3A, 0x41, 0x41, 0x21, 0x7A,
	0x3A, 0x42, 0x40, 0x20, 0x78,
	0x00, 0x9D, 0xA0, 0xA0, 0x7D,
	0x39, 0x44, 0x44, 0x44, 0x39,
	0x3D, 0x40, 0x40, 0x40, 0x3D,
	0x3C, 0x24, 0xFF, 0x24, 0x24,
	0x48, 0x7E, 0x49, 0x43, 0x66,
	0x2B, 0x2F, 0xFC, 0x2F, 0x2B,
	0xFF, 0x09, 0x29, 0xF6, 0x20,
	0xC0, 0x88, 0x7E, 0x09, 0x03,
	0x20, 0x54, 0x54, 0x79, 0x41,
	0x00, 0x00, 0x44, 0x7D, 0x41,
	0x30, 0x48, 0x48, 0x4A, 0x32,
	0x38, 0x40, 0x40, 0x22, 0x7A,
	0x00, 0x7A, 0x0A, 0x0A, 0x72,
	0x7D, 0x0D, 0x19, 0x31, 0x7D,
	0x26, 0x29, 0x29, 0x2F, 0x28,
	0x26, 0x29, 0x29, 0x29, 0x26,
	0x30, 0x48, 0x4D, 0x40, 0x20,
	0x38, 0x08, 0x08, 0x08, 0x08,
	0x08, 0x08, 0x08, 0x08, 0x38,
	0x2F, 0x10, 0xC8, 0xAC, 0xBA,
	0x2F, 0x10, 0x28, 0x34, 0xFA,
	0x00, 0x00, 0x7B, 0x00, 0x00,
	0x08, 0x14, 0x2A, 0x14, 0x22,
	0x22, 0x14, 0x2A, 0x14, 0x08,
	0xAA, 0x00, 0x55, 0x00, 0xAA,
	0xAA, 0x55, 0xAA, 0x55, 0xAA,
	0x00, 0x00, 0x00, 0xFF, 0x00,
	0x10, 0x10, 0x10, 0xFF, 0x00,
	0x14, 0x14, 0x14, 0xFF, 0x00,
	0x10, 0x10, 0xFF, 0x00, 0xFF,
	0x10, 0x10, 0xF0, 0x10, 0xF0,
	0x14, 0x14, 0x14, 0xFC, 0x00,
	0x14, 0x14, 0xF7, 0x00, 0xFF,
	0x00, 0x00, 0xFF, 0x00, 0xFF,
	0x14, 0x14, 0xF4, 0x04, 0xFC,
	0x14, 0x14, 0x17, 0x10, 0x1F,
	0x10, 0x10, 0x1F, 0x10, 0x1F,
	0x14, 0x14, 0x14, 0x1F, 0x00,
	0x10, 0x10, 0x10, 0xF0, 0x00,
	0x00, 0x00, 0x00, 0x1F, 0x10,
	0x10, 0x10, 0x10, 0x1F, 0x10,
	0x10, 0x10, 0x10, 0xF0, 0x10,
	0x00, 0x00, 0x00, 0xFF, 0x10,
	0x10, 0x10, 0x10, 0x10, 0x10,
	0x10, 0x10, 0x10, 0xFF, 0x10,
	0x00, 0x00, 0x00, 0xFF, 0x14,
	0x00, 0x00, 0xFF, 0x00, 0xFF,
	0x00, 0x00, 0x1F, 0x10, 0x17,
	0x00, 0x00, 0xFC, 0x04, 0xF4,
	0x14, 0x14, 0x17, 0x10, 0x17,
	0x14, 0x14, 0xF4, 0x04, 0xF4,
	0x00, 0x00, 0xFF, 0x00, 0xF7,
	0x14, 0x14, 0x14, 0x14, 0x14,
	0x14, 0x14, 0xF7, 0x00, 0xF7,
	0x14, 0x14, 0x14, 0x17, 0x14,
	0x10, 0x10, 0x1F, 0x10, 0x1F,
	0x14, 0x14, 0x14, 0xF4, 0x14,
	0x10, 0x10, 0xF0, 0x10, 0xF0,
	0x00, 0x00, 0x1F, 0x10, 0x1F,
	0x00, 0x00, 0x00, 0x1F, 0x14,
	0x00, 0x00, 0x00, 0xFC, 0x14,
	0x00, 0x00, 0xF0, 0x10, 0xF0,
	0x10, 0x10, 0xFF, 0x10, 0xFF,
	0x14, 0x14, 0x14, 0xFF, 0x14,
	0x10, 0x10, 0x10, 0x1F, 0x00,
	0x00, 0x00, 0x00, 0xF0, 0x10,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xF0, 0xF0, 0xF0, 0xF0, 0xF0,
	0xFF, 0xFF, 0xFF, 0x00, 0x00,
	0x00, 0x00, 0x00, 0xFF, 0xFF,
	0x0F, 0x0F, 0x0F, 0x0F, 0x0F,
	0x38, 0x44, 0x44, 0x38, 0x44,
	0x7C, 0x2A, 0x2A, 0x3E, 0x14,
	0x7E, 0x02, 0x02, 0x06, 0x06,
	0x02, 0x7E, 0x02, 0x7E, 0x02,
	0x63, 0x55, 0x49, 0x41, 0x63,
	0x38, 0x44, 0x44, 0x3C, 0x04,
	0x40, 0x7E, 0x20, 0x1E, 0x20,
	0x06, 0x02, 0x7E, 0x02, 0x02,
	0x99, 0xA5, 0xE7, 0xA5, 0x99,
	0x1C, 0x2A, 0x49, 0x2A, 0x1C,
	0x4C, 0x72, 0x01, 0x72, 0x4C,
	0x30, 0x4A, 0x4D, 0x4D, 0x30,
	0x30, 0x48, 0x78, 0x48, 0x30,
	0xBC, 0x62, 0x5A, 0x46, 0x3D,
	0x3E, 0x49, 0x49, 0x49, 0x00,
	0x7E, 0x01, 0x01, 0x01, 0x7E,
	0x2A, 0x2A, 0x2A, 0x2A, 0x2A,
	0x44, 0x44, 0x5F, 0x44, 0x44,
	0x40, 0x51, 0x4A, 0x44, 0x40,
	0x40, 0x44, 0x4A, 0x51, 0x40,
	0x00, 0x00, 0xFF, 0x01, 0x03,
	0xE0, 0x80, 0xFF, 0x00, 0x00,
	0x08, 0x08, 0x6B, 0x6B, 0x08,
	0x36, 0x12, 0x36, 0x24, 0x36,
	0x06, 0x0F, 0x09, 0x0F, 0x06,
	0x00, 0x00, 0x18, 0x18, 0x00,
	0x00, 0x00, 0x10, 0x10, 0x00,
	0x30, 0x40, 0xFF, 0x01, 0x01,
	0x00, 0x1F, 0x01, 0x01, 0x1E,
	0x00, 0x19, 0x1D, 0x17, 0x12,
	0x00, 0x3C, 0x3C, 0x3C, 0x3C,
	0x00, 0x00, 0x00, 0x00, 0x00,
}

var (
LCDWIDTH uint8 = 84
LCDHEIGHT uint8 = 48
)


// the memory buffer for the LCD
var pcd8544_buffer [84*48 / 8]byte = [84*48 / 8]byte{0,}

var pi_logo []byte  = []byte {
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0010 (16) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0xF8, 0xFC, 0xAE, 0x0E, 0x0E, 0x06, 0x0E, 0x06,   // 0x0020 (32) pixels
	0xCE, 0x86, 0x8E, 0x0E, 0x0E, 0x1C, 0xB8, 0xF0, 0xF8, 0x78, 0x38, 0x1E, 0x0E, 0x8E, 0x8E, 0xC6,   // 0x0030 (48) pixels
	0x0E, 0x06, 0x0E, 0x06, 0x0E, 0x9E, 0xFE, 0xFC, 0xF8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0040 (64) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0050 (80) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0060 (96) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x0F, 0x0F, 0xFE,   // 0x0070 (112) pixels
	0xF8, 0xF0, 0x60, 0x60, 0xE0, 0xE1, 0xE3, 0xF7, 0x7E, 0x3E, 0x1E, 0x1F, 0x1F, 0x1F, 0x3E, 0x7E,   // 0x0080 (128) pixels
	0xFB, 0xF3, 0xE1, 0xE0, 0x60, 0x70, 0xF0, 0xF8, 0xBE, 0x1F, 0x0F, 0x07, 0x00, 0x00, 0x00, 0x00,   // 0x0090 (144) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x00A0 (160) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x00B0 (176) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0xC0,   // 0x00C0 (192) pixels
	0xE0, 0xFC, 0xFE, 0xFF, 0xF3, 0x38, 0x38, 0x0C, 0x0E, 0x0F, 0x0F, 0x0F, 0x0E, 0x3C, 0x38, 0xF8,   // 0x00D0 (208) pixels
	0xF8, 0x38, 0x3C, 0x0E, 0x0F, 0x0F, 0x0F, 0x0E, 0x0C, 0x38, 0x38, 0xF3, 0xFF, 0xFF, 0xF8, 0xE0,   // 0x00E0 (224) pixels
	0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x00F0 (240) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0100 (256) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0110 (272) pixels
	0x00, 0x7F, 0xFF, 0xE7, 0xC3, 0xC1, 0xE0, 0xFF, 0xFF, 0x78, 0xE0, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0,   // 0x0120 (288) pixels
	0x60, 0x78, 0x38, 0x3F, 0x3F, 0x38, 0x38, 0x60, 0x60, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, 0xF8, 0x7F,   // 0x0130 (304) pixels
	0xFF, 0xE0, 0xC1, 0xC3, 0xE7, 0x7F, 0x3E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0140 (320) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0150 (336) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0160 (352) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x0F, 0x7F, 0xFF, 0xF1, 0xE0, 0xC0, 0x80, 0x01,   // 0x0170 (368) pixels
	0x03, 0x9F, 0xFF, 0xF0, 0xE0, 0xE0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xE0, 0xE0, 0xF0, 0xFF, 0x9F,   // 0x0180 (384) pixels
	0x03, 0x01, 0x80, 0xC0, 0xE0, 0xF1, 0x7F, 0x1F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x0190 (400) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x01A0 (416) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x01B0 (432) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,   // 0x01C0 (448) pixels
	0x03, 0x03, 0x07, 0x07, 0x0F, 0x1F, 0x1F, 0x3F, 0x3B, 0x71, 0x60, 0x60, 0x60, 0x60, 0x60, 0x71,   // 0x01D0 (464) pixels
	0x3B, 0x1F, 0x0F, 0x0F, 0x0F, 0x07, 0x03, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x01E0 (480) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,   // 0x01F0 (496) pixels
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 
	}

func my_setpixel(x uint8, y uint8,  color bool) {
	if ((x >= LCDWIDTH) || (y >= LCDHEIGHT)) {
		return
	}
	// x is which column
	if (color) {
		pcd8544_buffer[x+ (y/8)*LCDWIDTH] |= _BV(y%8)
	} else {
		pcd8544_buffer[x+ (y/8)*LCDWIDTH] &= ^(_BV(y%8))
	}
}

func LCDInit(SCLK,DIN,DC,CS,RST,BL,contrast uint8) (pin PCD8544_pin) {
	_din := DIN
	_sclk := SCLK
	_dc  := DC
	_rst  := RST
	_cs := CS
	_bl := BL
	_contrast := contrast

	pin = PCD8544_pin{	
       _din : _din,
	   _sclk : _sclk,
	   _dc : _dc,
	   _rst : _rst,
	   _cs: _cs,
	   _bl: _bl,
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
	if (_cs > 0) {
		csPin.Low()
	}

	rstPin.Low()
	
	//sleep 500ms
	_delay_ms(500)
    //time.Sleep(time.Duration(500) * time.Millisecond)
    rstPin.High()

	
	// get into the EXTENDED mode!
	pin.LCDCommand(PCD8544_FUNCTIONSET | PCD8544_EXTENDEDINSTRUCTION )

	// LCD bias select (4 is optimal?)
	pin.LCDCommand(PCD8544_SETBIAS | 0x4)

	// set VOP
	if (_contrast > 0x7f) {
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
     dcPin := rpio.Pin( pin._dc)
	 rpio.WritePin(dcPin,rpio.Low)
	 //dcPin.Low()
	 pin.LCDspiwrite(cmd)
}

func (pin PCD8544_pin) LCDspiwrite(cmd uint8) {
	csPin := rpio.Pin(pin._cs)
	dinPin := rpio.Pin(pin._din)
	sclkPin := rpio.Pin(pin._sclk)
	csPin.Low()
	ShiftOut(dinPin, sclkPin, MSBFIRST, cmd)
	csPin.High()


}

func (pin PCD8544_pin) LCDData(c uint8) {
	dcPin := rpio.Pin(pin._dc)
	dcPin.High()
	pin.LCDspiwrite(c)
}

func (pin PCD8544_pin) LCDSetcontrast(val uint8) {
	if val > 0x7f {
		val = 0x7f
	}
	pin.LCDCommand(PCD8544_FUNCTIONSET | PCD8544_EXTENDEDINSTRUCTION )
	pin.LCDCommand( PCD8544_SETVOP | val)
	pin.LCDCommand(PCD8544_FUNCTIONSET)
}

func (pin PCD8544_pin) LCDDisplay() {
	var col,maxcol,p uint8
	for p = 0 ; p < 6; p++ {
		pin.LCDCommand(PCD8544_SETYADDR | p)
		// start at the beginning of the row
		col = 0;
		maxcol = LCDWIDTH-1;
		pin.LCDCommand(PCD8544_SETXADDR | col);

		for ;col <= maxcol; col++ {
			//uart_putw_dec(col);
			//uart_putchar(' ');
			pin.LCDData(pcd8544_buffer[(LCDWIDTH*p)+col])
		}
	}
	pin.LCDCommand(PCD8544_SETYADDR )  // no idea why this is necessary but it is to finish the last byte?
}

func (pin PCD8544_pin) LCDShowRpiLogo() {
	var i uint8
	for i = 0; i < (LCDWIDTH*LCDHEIGHT)/8; i++ {
		pcd8544_buffer[i] = pi_logo[i]
	}
	pin.LCDDisplay()
}


func LCDClear() {
	var i uint8
	for i = 0; i < (LCDWIDTH*LCDHEIGHT)/8; i++ {
		pcd8544_buffer[i] = 0
	}
	cursor_y = 0
	cursor_x = 0
}


func isTrue(val uint8) bool {
	if val == 0 {
		//fmt.Printf("isTrue -> val=%d\n",val)
		return false
	}
	//fmt.Printf("isTrue -> val=%d\n",val)
	return true
}
func bool2RpiState(val bool) rpio.State {
	if val == true {
		return rpio.High
	}
	return rpio.Low
}

func ShiftOut(dinPin rpio.Pin,sclkPin rpio.Pin,bitOrder uint8,val uint8) {
	for i := 0; i < 8; i++ {
		if (bitOrder == LSBFIRST) {
			dinPin.Write(bool2RpiState(!!isTrue(val & (1 << i))))
		} else {
             dinPin.Write(bool2RpiState(!!isTrue(val & (1 << (7 - i)))))
		}

		sclkPin.Write(rpio.High)
		for j := CLKCONST_2; j >0; j-- {
			sclkPin.Write(rpio.Low)
		}
		
	}

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


func Str2ByteSlice(val string) []byte {
	return []byte(val)
}

func LCDDrawString(x uint8,y uint8, val []byte) {
	cursor_x = x
	cursor_y = y
	fmt.Printf("LCDDrawString -> val = %s\n",string(val))
	for i := 0; i < len(val); i++ {
		LCDWrite(val[i])
	}
}

func LCDWrite(c byte) {
      
	if c == '\n' {
       cursor_y += textsize*8
	   cursor_x = 0
	} else if c == '\r' {
		//skip em
	} else {
       LCDDrawchar(cursor_x,cursor_y,c)
	   cursor_x += textsize*6
	   if (cursor_x >= (LCDWIDTH -5)) {
		   cursor_x = 0
		   cursor_y += 8
	   }
	   if (cursor_y >= LCDHEIGHT) {
		   cursor_y = 0

	   }
	}
}

func LCDDrawchar(x uint8,y uint8, c byte) {
       if y >= LCDHEIGHT {
		   return
	   }
	   if x + 5 >= LCDWIDTH {
		   return
	   }
	   var i,j uint8
	   for i = 0; i < 5; i++ {
		   var d byte = (font[(c*5) +i])
		   //fmt.Printf("LCDDrawchar -> c=%d,d=%d\n",c,d)
		   for j =0; j< 8; j++ {
			   if (isTrue(d & _BV(j))) {
				   my_setpixel(x+i,y+j,textcolor)
			   } else {
				my_setpixel(x+i,y+j,!textcolor)
			   }
		   }
	   }

	   for j = 0 ; j < 8; j++ {
		my_setpixel(x+5, y+j, !textcolor)
	   }
	   //updateBoundingBox(x, y, x+5, y + 8)

}


func main() {
	
	//define the gpio pin for pcd8544
	//pin setup
var (
	SCLK uint8 = 17
	DIN uint8 = 18
	DC uint8 = 27
	CS uint8 = 22
	RST uint8 = 23
	BL uint8 = 4
)

	var contrast uint8 = 45


	fmt.Printf("Raspberry Pi Nokia5110 sysinfo display\n")
	//init gpio in main func

	
	//Init LCD
	pin := LCDInit(SCLK,DIN,DC,CS,RST,BL,contrast)

	LCDClear()
	pin.LCDShowRpiLogo()

	for ;; {
		LCDClear()

		
		timeObj := time.Now()

		//month := timeObj.Month()
		//monthstr := Month2String(MonthType(month))
	
		//day := timeObj.Day()
	
		hour := timeObj.Hour()
		minute := timeObj.Minute()
		second := timeObj.Second()
	
		timeInfo := fmt.Sprintf("%d:%d:%d",hour,minute,second)
		timeInfoBytes := []byte(timeInfo)

		LCDDrawString(0 ,0,timeInfoBytes)
		

		//pin.LCDShowRpiLogo()

		pin.LCDDisplay()

		time.Sleep(1000)

	}

}


