package parsepcap

func decodeBlocks(mergedPoints *[]Point, ip4Channel *iterationInfo, nextPacketData *[]byte, workerIndex uint8, totalWorkers uint8,
	endFrame *int, startFrame *int) {
	productID := getProductID(&((*ip4Channel).currPacketData))

	for colIndex := uint8(0); colIndex < 12; colIndex++ {
		// firingTime := getTime(&((*ip4Channel).currPacketData))
		currAzimuth, nextAzimuth := getCurrentAndNextRawAzimuths(&((*ip4Channel).currPacketData), nextPacketData, colIndex)

		// Check if new frame
		if currAzimuth > nextAzimuth && currAzimuth > 35900 && nextAzimuth < 150 {
			// if firingTime-*prevFiringTime >= 55296 {
			(*ip4Channel).isFinished = ((*ip4Channel).frameCount >= *endFrame) && (*endFrame > 0)
			if (*ip4Channel).isFinished {
				return
			}

			if len((*ip4Channel).currPoints) > 0 {
				// Do something with the points
				(*ip4Channel).isReady = true
				// fmt.Println(productID, workerIndex, (*ip4Channel).frameCount, len((*ip4Channel).currPoints), firingTime, currAzimuth, nextAzimuth, (*ip4Channel).isReady)
			}

			(*ip4Channel).frameCount++
			// (*ip4Channel).currPoints = nil
		}

		// Check if frame will be assigned to the worker, otherwise skip the packet
		isDecode := (*ip4Channel).frameCount%int(totalWorkers) == int(workerIndex) && (*ip4Channel).frameCount >= *startFrame
		if !isDecode {
			continue
		}

		for rowIndex := uint8(0); rowIndex < 32; rowIndex++ {
			// Check if distance is non zero, skip point if distance is zero
			distance, reflectivity := getDistanceAndReflectivity(&((*ip4Channel).currPacketData), colIndex, rowIndex)
			if distance == 0 {
				continue
			}

			azimuth := getPrecisionAzimuth(currAzimuth, nextAzimuth, rowIndex, productID)
			X, Y, Z := getXYZCoordinates(&distance, azimuth, productID, rowIndex)
			laserID := getLaserID(productID, rowIndex)

			timeStamp := getTimeStamp(&((*ip4Channel).currPacketData), rowIndex, colIndex)

			newPoint := Point{Distance: distance, LidarModel: productID,
				X: X, Y: Y, Z: Z, Intensity: reflectivity,
				Azimuth: azimuth, LaserID: laserID, Timestamp: timeStamp}

			// If there is a new frame, save the points to the next frame
			if (*ip4Channel).isReady {
				(*ip4Channel).nextPoints = append((*ip4Channel).nextPoints, newPoint)
			} else {
				(*ip4Channel).currPoints = append((*ip4Channel).currPoints, newPoint)
			}
		}
	}
}
