package core

func expireSample() float32 {
	var limit = 20
	var deletedKeys = 0

	for key, value := range store {
		if value.ValidTill != -1 {
			limit--

			if value.HasExpired() {
				delete(store, key)
				deletedKeys++
			}
		}
		if limit == 0 {
			break
		}
	}

	return float32(deletedKeys) / float32(20)
}

func SafeDeleteExpiredKeys() {
	for {
		frac := expireSample()
		if frac < 0.25 {
			break
		}
	}
}
