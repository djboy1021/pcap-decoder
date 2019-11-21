package pcapparser

import (
	"encoding/binary"
	"sync"
)

func setChannel(data *[]byte, channels *[]LidarChannel, chIndex uint8, blkIndex uint16, wg *sync.WaitGroup) {
	(*channels)[chIndex] = LidarChannel{
		Distance:     uint16((*data)[blkIndex+1])<<8 + uint16((*data)[blkIndex]),
		Reflectivity: (*data)[blkIndex+2]}

	wg.Done()
}

func setChannels(data *[]byte, index uint16, channels *[]LidarChannel, blocksWG *sync.WaitGroup) {
	var channelsWaitGroup sync.WaitGroup
	chIndex := uint8(0)
	channelsWaitGroup.Add(32)
	for blkIndex := index + 4; blkIndex < index+100; blkIndex += 3 {
		setChannel(data, channels, chIndex, blkIndex, &channelsWaitGroup)
		chIndex++
	}
	channelsWaitGroup.Wait()

	blocksWG.Done()
}

func setBlocks(data *[]byte, blocks *[]LidarBlock) {
	blkIndex := uint8(0)
	var blocksWG sync.WaitGroup
	blocksWG.Add(12)
	for index := uint16(42); index < 1242; index += 100 {
		// Set Azimuth
		(*blocks)[blkIndex].Azimuth = binary.LittleEndian.Uint16((*data)[index+2 : index+4])

		// Initialize channels
		(*blocks)[blkIndex].Channels = make([]LidarChannel, 32)

		setChannels(data, index, &((*blocks)[blkIndex].Channels), &blocksWG)
		blkIndex++
	}
	blocksWG.Wait()
}
