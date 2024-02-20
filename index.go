package gzran

import (
	"encoding/gob"
	"io"
	"sort"
)

// Index collects decompressor state at offset Points.
// gzseek.Reader adds points to the index on the fly as decompression proceeds.

type Index struct {
	Digest       uint32 // CRC-32, IEEE polynomial (section 8)
	Size         uint32 // Uncompressed size (section 2.3.1)
	FurthestRead int64
	DigestDone   bool
	Points       []Point
}

func NewIndex(compressedOffset int64) Index {
	return Index{
		Points: []Point{{
			CompressedOffset: compressedOffset,
		},
		},
	}
}

// LoadIndex deserializes an Index from the given io.Reader.
func LoadIndex(r io.Reader) (Index, error) {
	dec := gob.NewDecoder(r)
	var idx Index
	err := dec.Decode(&idx)
	return idx, err
}

// WriteTo serializes the index to the given io.Writer.
// It can be deserialized with LoadIndex.
func (idx Index) WriteTo(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(idx)
}

func (idx Index) lastUncompressedOffset() int64 {
	if len(idx.Points) == 0 {
		return 0
	}

	return idx.Points[len(idx.Points)-1].UncompressedOffset
}

func (idx Index) closestPointBefore(offset int64) Point {
	j := sort.Search(len(idx.Points), func(j int) bool {
		return idx.Points[j].UncompressedOffset > offset
	})

	if j == 0 {
		return Point{}
	}

	return idx.Points[j-1]
}

// Point holds the decompressor state at a given offset within the uncompressed data.
type Point struct {
	CompressedOffset   int64
	UncompressedOffset int64
	DecompressorState  []byte
}
