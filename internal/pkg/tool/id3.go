package tool

import (
	"encoding/binary"
	"os"
)

func encodeSyncSafe(size int) uint32 {
	return uint32((size&0x7F)<<21 | (size&0x7F00)<<14 | (size&0x7F0000)<<7 | (size&0x7F000000)>>0)
}

func createID3Header(size int) []byte {
	header := make([]byte, 10)
	copy(header[0:3], "ID3")                                     // Identifier
	header[3] = 0x04                                             // Version: ID3v2.4
	header[4] = 0x00                                             // Revision
	header[5] = 0x00                                             // Flags (no extended header)
	binary.BigEndian.PutUint32(header[6:], encodeSyncSafe(size)) // Tag size
	return header
}

func createTextFrame(frameID, content string) []byte {
	contentBytes := append([]byte(content), 0x00) // Add null terminator
	size := len(contentBytes) + 1                 // 1 byte for encoding
	frame := make([]byte, 10+size)
	copy(frame[0:4], frameID)                                    // Frame ID
	binary.BigEndian.PutUint32(frame[4:8], encodeSyncSafe(size)) // Sync-safe encoded frame size
	frame[8] = 0x00                                              // Flags
	frame[9] = 0x00
	frame[10] = 0x03               // UTF-8 encoding
	copy(frame[11:], contentBytes) // Frame content
	return frame
}

func MakeId3Tag(title, artist string) []byte {
	titleFrame := createTextFrame("TIT2", title)
	artistFrame := createTextFrame("TPE1", artist)
	tagSize := len(titleFrame) + len(artistFrame)
	header := createID3Header(tagSize)
	tag := append(header, titleFrame...)
	return append(tag, artistFrame...)
}

func WriteID3Tag(filePath, title, artist string) error {
	originalData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	tag := MakeId3Tag(title, artist)
	newData := append(tag, originalData...)
	return os.WriteFile(filePath, newData, 0o644)
}
