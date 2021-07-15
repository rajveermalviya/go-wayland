package xcursor

import (
	"encoding/binary"
	"io"
	"os"
)

// Cursor files start with a header.  The header
// contains a magic number, a version number and a
// table of contents which has type and offset information
// for the remaining tables in the file.
//
// File minor versions increment for compatible changes
// File major versions increment for incompatible changes (never, we hope)
//
// Chunks of the same type are always upward compatible.  Incompatible
// changes are made with new chunk types; the old data can remain under
// the old type.  Upward compatible changes can add header data as the
// header lengths are specified in the file.
//
// File:
//  FileHeader
//  LISTofChunk
//
// FileHeader:
//  CARD32           magic      magic number
//  CARD32           header     bytes in file header
//  CARD32           version    file version
//  CARD32           ntoc       number of toc entries
//  LISTofFileToc    toc        table of contents
//
// FileToc:
//  CARD32    type        entry type
//  CARD32    subtype     entry subtype (size for images)
//  CARD32    position    absolute file position

const xcursorMagic = 0x72756358 // "Xcur" LSBFirst

const (
	fileHeaderLen = (4 * 4)
	fileTocLen    = (3 * 4)
)

type fileToc struct {
	typ      uint32 // chunk type
	subtype  uint32 // subtype (size for images)
	position uint32 // absolute position in file
}

type fileHeader struct {
	magic   uint32    // magic number
	header  uint32    // byte length of header
	version uint32    // file version number
	ntoc    uint32    // number of toc entries
	tocs    []fileToc // table of contents
}

// The rest of the file is a list of chunks, each tagged by type
// and version.
//
// Chunk:
//  ChunkHeader
//  <extra type-specific header fields>
//  <type-specific data>
//
// ChunkHeader:
//  CARD32    header     bytes in chunk header + type header
//  CARD32    type       chunk type
//  CARD32    subtype    chunk subtype
//  CARD32    version    chunk type version

const chunkHeaderLen = (4 * 4)

type chunkHeader struct {
	header  uint32 // bytes in chunk header
	typ     uint32 // chunk type
	subtype uint32 // chunk subtype (size for images)
	version uint32 // version of this type
}

// Each cursor image occupies a separate image chunk.
// The length of the image header follows the chunk header
// so that future versions can extend the header without
// breaking older applications
//
// Image:
// ChunkHeader     header    chunk header
// CARD32          width     actual width
// CARD32          height    actual height
// CARD32          xhot      hotspot x
// CARD32          yhot      hotspot y
// CARD32          delay     animation delay
// LISTofCARD32    pixels    ARGB pixels

const (
	imageType      = 0xfffd0002
	imageVersion   = 1
	imageHeaderLen = (chunkHeaderLen + (5 * 4))
	imageMaxSize   = 0x7fff // 32767x32767 max cursor size
)

type Image struct {
	Version  uint32 // version of the image data
	Size     uint32 // nominal size for matching
	Width    uint32 // actual width
	Height   uint32 // actual height
	HotspotX uint32 // hot spot x (must be inside image)
	HotspotY uint32 // hot spot y (must be inside image)
	Delay    uint32 // animation delay to next frame (ms)
	Pixels   []byte // pointer to pixels
}

func readUint32(f *os.File) (uint32, bool) {
	if f == nil {
		return 0, false
	}

	bs := make([]byte, 4)
	if _, err := io.ReadFull(f, bs); err != nil {
		return 0, false
	}

	return binary.LittleEndian.Uint32(bs), true
}

func readFileHeader(f *os.File) (fh fileHeader, ok bool) {
	if f == nil {
		return fh, false
	}

	magic, ok := readUint32(f)
	if !ok {
		return fh, false
	}
	if magic != xcursorMagic {
		return fh, false
	}

	header, ok := readUint32(f)
	if !ok {
		return fh, false
	}
	version, ok := readUint32(f)
	if !ok {
		return fh, false
	}
	ntoc, ok := readUint32(f)
	if !ok {
		return fh, false
	}

	skip := header - fileHeaderLen
	if skip == 0 {
		_, err := f.Seek(int64(skip), os.SEEK_CUR)
		if err != nil {
			return fh, false
		}
	}

	if ntoc > 0x10000 {
		return fh, false
	}

	fh.magic = magic
	fh.header = header
	fh.version = version
	fh.ntoc = ntoc
	fh.tocs = make([]fileToc, ntoc)

	for n := uint32(0); n < ntoc; n++ {
		typ, ok := readUint32(f)
		if !ok {
			return fh, false
		}
		subtype, ok := readUint32(f)
		if !ok {
			return fh, false
		}
		position, ok := readUint32(f)
		if !ok {
			return fh, false
		}

		fh.tocs[n] = fileToc{
			typ:      typ,
			subtype:  subtype,
			position: position,
		}
	}

	return fh, true
}

func seekToToc(f *os.File, fh fileHeader, toc int) bool {
	if f == nil {
		return false
	}

	if _, err := f.Seek(int64(fh.tocs[toc].position), io.SeekStart); err != nil {
		return false
	}

	return true
}

func readChunkHeader(f *os.File, fh fileHeader, toc int) (ch chunkHeader, ok bool) {
	if f == nil {
		return ch, false
	}

	ok = seekToToc(f, fh, toc)
	if !ok {
		return ch, false
	}

	header, ok := readUint32(f)
	if !ok {
		return ch, false
	}
	typ, ok := readUint32(f)
	if !ok {
		return ch, false
	}
	subtype, ok := readUint32(f)
	if !ok {
		return ch, false
	}
	version, ok := readUint32(f)
	if !ok {
		return ch, false
	}

	// sanity check
	if typ != fh.tocs[toc].typ || subtype != fh.tocs[toc].subtype {
		return ch, false
	}

	ch.header = header
	ch.typ = typ
	ch.subtype = subtype
	ch.version = version

	return ch, true
}

func dist(a, b uint32) uint32 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func findBestSize(fh fileHeader, size uint32) (bestSize uint32, nsizes int) {
	for n := uint32(0); n < fh.ntoc; n++ {
		if fh.tocs[n].typ != imageType {
			continue
		}

		thisSize := fh.tocs[n].subtype
		if bestSize == 0 || dist(thisSize, size) < dist(bestSize, size) {
			bestSize = thisSize
			nsizes = 1
		} else if thisSize == bestSize {
			nsizes++
		}
	}

	return bestSize, nsizes
}

func findImageToc(fh fileHeader, size uint32, count int) int {
	toc := uint32(0)
	for ; toc < fh.ntoc; toc++ {
		if fh.tocs[toc].typ != imageType {
			continue
		}

		thisSize := fh.tocs[toc].subtype
		if thisSize != size {
			continue
		}
		if count == 0 {
			break
		}
		count--
	}
	if toc == fh.ntoc {
		return -1
	}
	return int(toc)
}

func readImage(f *os.File, fh fileHeader, toc int) (img Image, ok bool) {
	if f == nil {
		return img, false
	}

	ch, ok := readChunkHeader(f, fh, toc)
	if !ok {
		return img, false
	}

	width, ok := readUint32(f)
	if !ok {
		return img, false
	}
	height, ok := readUint32(f)
	if !ok {
		return img, false
	}
	xhot, ok := readUint32(f)
	if !ok {
		return img, false
	}
	yhot, ok := readUint32(f)
	if !ok {
		return img, false
	}
	delay, ok := readUint32(f)
	if !ok {
		return img, false
	}

	// sanity check data
	if width > imageMaxSize || height > imageMaxSize {
		return img, false
	}
	if width <= 0 || height <= 0 {
		return img, false
	}
	if xhot > width || yhot > height {
		return img, false
	}

	// Create image
	size := height
	if width > height {
		size = width
	}

	version := uint32(imageVersion)
	if ch.version < version {
		version = ch.version
	}

	pixLen := 4 * width * height
	pixRGBA := make([]byte, pixLen)
	if _, err := io.ReadFull(f, pixRGBA); err != nil {
		return img, false
	}

	img.Version = version
	img.Size = size
	img.Width = width
	img.Height = height
	img.HotspotX = xhot
	img.HotspotY = yhot
	img.Delay = delay
	img.Pixels = pixRGBA

	return img, true
}

func fileLoadImages(f *os.File, size int) []Image {
	if f == nil {
		return nil
	}

	if size < 0 {
		return nil
	}

	fh, ok := readFileHeader(f)
	if !ok {
		return nil
	}

	bestSize, nsize := findBestSize(fh, uint32(size))
	if bestSize == 0 {
		return nil
	}

	images := make([]Image, nsize)

	for n := 0; n < nsize; n++ {
		toc := findImageToc(fh, bestSize, n)
		if toc < 0 {
			return nil
		}

		image, ok := readImage(f, fh, toc)
		if !ok {
			return nil
		}

		images[n] = image
	}

	return images
}
