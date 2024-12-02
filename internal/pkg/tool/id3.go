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
	copy(header[0:3], "ID3")            // Identifier
	header[3] = 0x04                   // Version: ID3v2.4
	header[4] = 0x00                   // Revision
	header[5] = 0x00                   // Flags (no extended header)
	binary.BigEndian.PutUint32(header[6:], encodeSyncSafe(size)) // Tag size
	return header
}

func createTextFrame(frameID, content string) []byte {
	contentBytes := []byte(content)
	size := len(contentBytes) + 1 // 1 byte for encoding
	frame := make([]byte, 10+size)
	copy(frame[0:4], frameID)                     // Frame ID
	binary.BigEndian.PutUint32(frame[4:8], uint32(size)) // Frame size
	frame[8] = 0x00                               // Flags
	frame[9] = 0x00
	frame[10] = 0x03                              // UTF-8 encoding
	copy(frame[11:], contentBytes)                // Frame content
	return frame
}

func WriteID3Tag(filePath string, title, artist string) error {
	// Read the original file
	originalData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Create frames
	titleFrame := createTextFrame("TIT2", title)
	artistFrame := createTextFrame("TPE1", artist)

	// Calculate total tag size
	tagSize := len(titleFrame) + len(artistFrame)
    syncSafeSize := encodeSyncSafe(tagSize)

	// Create header using syncSafeSize
	header := createID3Header(int(syncSafeSize))

	// Create header
	//header := createID3Header(tagSize)

	// Create the full tag
	tag := append(header, titleFrame...)
	tag = append(tag, artistFrame...)

	// Write the new file
	err = os.WriteFile(filePath, append(tag, originalData...), 0644)
	if err != nil {
		return err
	}

	return nil
}
