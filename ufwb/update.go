// TODO Change the Validate to collect errors, instead of stopping on first
// TODO Use validationError as much as possible
package ufwb

import (
	"bramp.net/dsector/toerr"
	"fmt"
	"strconv"
)

func (u *Ufwb) Get(id string) (Element, bool) {
	if _, err := strconv.Atoi(id); err == nil {
		id = "id:" + id
	}

	e, found := u.Elements[id]
	return e, found
}

func (u *Ufwb) GetScript(id string) (*Script, bool) {
	if _, err := strconv.Atoi(id); err == nil {
		id = "id:" + id
	}

	e, found := u.Scripts[id]
	return e, found
}

func (g *Grammar) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {

	if e, found := u.Get(g.Xml.Start); found {
		if s, ok := e.(Element); ok {
			g.Start = s
		} else {
			errs.Append(&validationError{e: g, err: fmt.Errorf("start element %q is not a Element", g.Xml.Start)})
		}
	} else {
		errs.Append(&validationError{e: g, err: fmt.Errorf("start element %q not found", g.Xml.Start)})
	}
}

func (s *Structure) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {

	// Repeat:[ id:1001 id:101 id:106 id:1064 id:1071 id:1078 id:1085 id:1092 id:1098 id:1105 id:115 id:1204 id:157 id:164 id:171 id:1730 id:1758 id:18 id:180 id:194 id:199 id:221 id:230 id:24 id:243 id:25 id:253 id:254 id:264 id:269 id:285 id:29 id:295 id:320 id:322 id:378 id:391 id:5259 id:5272 id:5287 id:5478 id:5620 id:5634 id:58 id:5824 id:5969 id:5984 id:6 id:6172 id:62 id:6313 id:6328 id:64 id:6516 id:691 id:692 id:72 id:7630 id:7638 id:7664 id:7689 id:7690 id:7770 id:8088 id:820 id:829 id:8328 id:838 id:8413 id:847 id:856 id:86 id:861 id:868 id:8768 id:8781 id:8790 id:8812 id:8857 id:8875 id:8881 id:8891 id:8897 id:8920 id:8926 id:8932 id:8935 id:9020 id:9040 id:9081 id:9094 id:9103 id:9125 id:9154 id:9170 id:9188 id:9194 id:9204 id:9210 id:9233 id:9239 id:9245 id:9248 id:9264 id:9333 id:934 id:935 id:9353 id:947 id:960 id:961 id:973 id:974 id:989]
	// RepeatMax:[ -1 100 100000 110 127 16 2147483647 256 40 6 600 99 ClassSize Count CycleCount FieldCount FrameCount Height/2 MarkerCount MethodSize NumberOfBlocks NumberOfRecords NumberOfRecords-1 NumberOfSections OpcodeSize RecordCount StyleCount ValueCount numberOfHMetrics unlimited]
	// ValueExpression:[ Address Application Body Comment Date Description EUI Address Filter FontStyleCode Hardware Header MAC Address Name NameString Number OS PacketData Speed String TimeZone Value VirtualAddress]
	// Length:[ (ContentLength+4)*2 (ipart((Width+3)/4))*4 (ipart((Width+31)/32))*4 0 1 10 100 1014 1024 10240 1032 1045 10769 1084 10864 11016 1116 112 1140 1152 1153940 116 11604 11832 12 1226 1244 1248 1272 128 12885 1304 132 1320 135683 136 137757 1390 139181 14 + descriptionLength 1400 14164 142 142964 14300 1430564 143420 1450 1480060 1499852 151496 1529816 15524 1556 156 158626 16 160 16256 16320 16338 16384 1652 168 16932 17 1704 176 1760 1776 18 180 180606 1824 184 184716 18527492 190 192 1920 19221 196590 1997636 1997748 2 20 200 2035288 2048 2066 2096 2100 2147483647 219560 22 220 22477 226 2265948 2266 2276 2312424 2312460 2319 2384 24 2412 243 2466 2480 252 2528 256 260 2600167 27 272 27586 278 28 28085 28252 28364 296220 3 3022 312 316 32 3232 328 32844 3296 331 332 3340 336 344 34785 348 3508 3540 3568 36 37016 3732 38 3800 3941 3950 39957 4 40 400 40032 404 40710 408 4116 420 424 4328 436 44 440 4472 45072 452 4532 4615 47 476495 4779 48 484 48576 5 50 504 505 5100 512 52 520 5408 54936 572 576 5860 592 6 604 610 62337 6316 64 65536 656 680 6818 7 72 7643 7700 8 80 803 804 822 868 89008 892 8932 9 90796 91120 916 93092 934 9444 952 96 992974 BlockSize ChallengeDataLength DLCDataLength DataLength Document_Size Extra_field_length File_comment_length HeaderLength Length Length + 6 PartContentLength+4 PartLength PayloadLength PositionOfFirstDataRecord PrivateSize REC_LEN+4 RGLength RecordLength Remaining SectionLength Size Size + 6 SizeOfOptionalHeader StylesLength Tlength TotalLength-12 ValueCount ceil(dt_size/page_size)*page_size ceil(kernel_size/page_size)*page_size ceil(ramdisk_size/page_size)*page_size ceil(second_size/page_size)*page_size frame.cap_len length nSize page_size prev.BodyLength prev.Compressed_size prev.DataLength prev.ExtraFormatBytes prev.FieldNameLength prev.FieldValueLength prev.Length prev.RGLength prev.RgnDataSize prev.RgnSize prev.cbData prev.cbName prev.chunkSize prev.size prev.sizeofcmds size this.Length this.RGLength this.Record Length this.SFLength this.Size this.biSize this.cmdsize totalChunkSize-16]
	// Encoding:[ IBM500 IBM850 ISCII,version=0 ISO_8859-1:1987 UTF-16 UTF-8 macintosh]
	// Floating:[ yes]
	// Order:[ variable]
	// Alignment:[ 0 1 1024 2 4 8]
	// ConsistsOf:[ ChunkReference File File Header Tag Table Element Timestamp id:1057 id:11 id:135 id:1372 id:154 id:252 id:312 id:36 id:3678 id:46 id:470 id:55 id:6423 id:6432 id:68 id:8114]
	// RepeatMin:[ 0 110 16 255 40 FieldCount MarkerCount]
	// LengthOffset:[ 1 2]
	// Endian:[ big dynamic little]
	// Signed:[ no yes]]
	s.parent = parent

	// TODO Add Min/Max
	//if s.Order() == FixedOrder {
	// TODO if FixedOrder then Min/Max should equal # of children
	//}
}

func (n *Number) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {

	// Length:[ 1 11 12 13 16 2 20 24 2^Count 2^NumberOfBytes 3 32 4 5 6 64 7 8 offsetRefSize]
	// LengthUnit:[ bit]
	// Endian:[ big dynamic little]
	// Signed:[ no yes]
	// Type:[ float integer]]
	// MustMatch:[ yes]
	// RepeatMax:[ -1 0 10 101 109 11 110 127 130 15 153 16 18 19 194 2 202 21 22 251 256 3 30 32 33 347 354 355 36 386 4 411 412 42 440 45 48 54 58 60 63 69 7 75 78 8 9 BaseNumber FrameCount LineInfoLength NumParts NumPoints ObjectCount ValueCount Width numberOfGlyphs segCountX2 / 2]
	// RepeatMin:[ 0 10 101 109 11 110 127 130 15 153 16 18 19 194 2 202 21 22 251 256 3 30 33 347 354 355 36 386 4 411 412 42 440 45 48 54 58 60 63 69 7 75 78 8 9 BaseNumber LineInfoLength ValueCount]]
	// ValueExpression:[ FontAssociationCount+1]
	// MinVal:[ -128 -32768 0 0x0 0xA0D0A 0xA0D0A00 0xA0D0D 0xC1 0xD0D0A00 1 10 1024 129 14 15790321 16 2 24 3 32768 4 5 61680 64 8 9]
	// MaxVal:[ -1 0 0x1 0xA0D0AFF 0xD0D0AFF 0xF8 0xFF 0xFF0A0D0A 0xFF0A0D0D 0xFFFE 1 10 100 1023 12 127 15 15988470 16 17 19 191 2 20 2147483647 215 23 250 254 255 3 300 31 32386 32767 35 4 4000 4294967295 59 62195 62969 63 63993 65534 65535 65536 7 70 8 8388607 9 96 99 999]
	n.parent = parent

	for _, v := range n.values {
		bs, err := parseInt(v.Xml.Value, 0, 0, n.Signed())
		if err != nil {
			errs.Append(err)
		}
		v.value = bs
	}

	// TODO Check Masks  []*Mask       `xml:"mask,omitempty"`
}

func (b *Binary) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {
	// Length:[ 0 0x3F2 - PayloadLength 0xBF - ServerStringLength 0xFF - FilenameStringLength 1 10 1024 10520 11 1144 12 1276 128 13 1344 13628 13656 13988 14 144 1448 1476 15 15104 155 156 16 17336 18088 184 190 1960 2 22 228 24 24924 26280 272 28 28054 29252 3 3140 316 32 3468 37 3700 376 38 38911 3976 39963 4 40 400 404 42 435 44 4432 459 473 48 497 5 50 512 52416 53644 544 564 578756 58019 6 60928 6144 64 6428 68 7 70239 70767 72 772 7956 8 808 852 8766 888 8894 8898 9 908 BitsPerPixel/8 ByteCount DataSize FieldLength FileSize FilenameStringLength Frame_length_with_hdr-7 Frame_length_with_hdr-9 HeaderExtentionLength Length Length - 192 Length -128 NALULength NumberOfBytes+1 PacketLength SampleSize*SampleNumber ServerStringLength Size ValueCount cbSignature dt_size kernel_size ramdisk_size remaining second_size select(mod(FileSize, 512) - 1, 0, 512 - mod(FileSize, 512), 512 - mod(FileSize, 512)) size - 6]
	// LengthUnit:[ bit]
	// RepeatMin:[ 0 8 BaseNumber]
	// RepeatMax:[ -1 -PackbitCode + 1 40 8 BaseNumber prev.Frames]
	// MustMatch:[ no yes]]

	//b.unused = yesno(b.Xml.Unused, errs)
	b.parent = parent
}

func (s *String) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {

	// Length:[ 0 1 10 100 11 114 12 128 13 14 16 16256 17 2 256 28 3 30 32 4 44 46 512 6 60 64 7 8 80 CharCount CharCount*2 CommentLength CommentsSize File_name_length Length NameSize NumberOfChars Remaining ShipID2Length ShipIDLength Size StringLength * 2 ValueCount descriptionLength length length - 1 nameLength raceLength remaining]
	// Type:[fixed-length pascal zero-terminated]]
	// Encoding:[ ANSI_X3.4-1968 IBM850 ISO_8859-1:1987 Shift_JIS UTF-16 UTF-16BE UTF-16LE UTF-7 UTF-8 macintosh]
	// MustMatch:[ yes]
	// RepeatMin:[ 0 107 18 3 355 412 58 66 78 BaseNumber]
	// RepeatMax:[ 100 107 18 3 355 412 58 66 78 BaseNumber numberOfGlyphs]

	s.parent = parent

	/*
		switch s.Typ() {
		//case "fixed-length":
			//if s.Length() == "" {
			//	errs.Append(&validationError{e: s, err: errors.New("fixed-length strings requires a length")})
			//}
		case "pascal":
			// TODO I don't think this is strictly required, I just don't know how to handle them yet!
			//	errs.Append(&validationError{e: s, err: errors.New("pascal strings requires a length")})
		}
	*/

	// TODO Check the FixedValues match the length
	// TODO Check Values []*FixedValue `xml:"fixedvalue,omitempty"`
}

func (c *Custom) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {
	if s, found := u.GetScript(c.Xml.Script); found {
		c.script = s
	} else {
		errs.Append(&validationError{e: c, err: fmt.Errorf("script %q not found", c.Xml.Script)})
	}
}

func (g *GrammarRef) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {
	// TODO Load the grammar
}

func (o *Offset) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {
	o.parent = parent

	if o.Xml.RelativeTo != "" {
		if e, found := u.Get(o.Xml.RelativeTo); found {
			o.relativeTo = e
		} else {
			errs.Append(&validationError{e: o, err: fmt.Errorf("relativeTo %q not found", o.Xml.RelativeTo)})
		}
	}

	if o.Xml.References != "" {
		if e, found := u.Get(o.Xml.References); found {
			o.references = e
		} else {
			errs.Append(&validationError{e: o, err: fmt.Errorf("references %q not found", o.Xml.References)})
		}
	}

	if o.Xml.ReferencedSize != "" {
		if e, found := u.Get(o.Xml.ReferencedSize); found {
			o.referencedSize = e
		} else {
			errs.Append(&validationError{e: o, err: fmt.Errorf("referencedSize %q not found", o.Xml.ReferencedSize)})
		}
	}
}

func (s *Script) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {
	// TODO Check we support the language
	// TODO Actually parse the language at this point
	// TODO Finish Script
}

func (s *StructRef) update(u *Ufwb, parent *Structure, errs *toerr.Errors) {
	ref := s.Xml.Structure
	if ref == "" {
		// Old versions used the name field, instead of a structure field
		ref = s.Xml.Name
	}

	if e, found := u.Get(ref); found {
		if structure, ok := e.(*Structure); ok {
			s.structure = structure
		} else {
			errs.Append(&validationError{e: s, err: fmt.Errorf("reference element %q is not a structure", ref)})
		}

	} else {
		errs.Append(&validationError{e: s, err: fmt.Errorf("referenced struct %q not found", ref)})
	}
}
