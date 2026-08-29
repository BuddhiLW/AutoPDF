package composition

// DirtySet compares two manifests and propagates changed component identities
// through reverse dependency edges.
func DirtySet(previous, next RenderManifest) []string {
	previousHashes := make(map[string]string, len(previous.Fragments))
	for _, fragment := range previous.Fragments {
		previousHashes[fragment.ComponentID] = fragment.Hash
	}
	nextHashes := make(map[string]string, len(next.Fragments))
	dirty := make(map[string]struct{})
	for _, fragment := range next.Fragments {
		nextHashes[fragment.ComponentID] = fragment.Hash
		if previousHashes[fragment.ComponentID] != fragment.Hash {
			dirty[fragment.ComponentID] = struct{}{}
		}
	}
	for componentID := range previousHashes {
		if _, exists := nextHashes[componentID]; !exists {
			dirty[componentID] = struct{}{}
		}
	}

	changed := true
	for changed {
		changed = false
		for componentID, dependencies := range next.Dependencies {
			if _, alreadyDirty := dirty[componentID]; alreadyDirty {
				continue
			}
			for _, dependency := range dependencies {
				if _, dependencyDirty := dirty[dependency]; dependencyDirty {
					dirty[componentID] = struct{}{}
					changed = true
					break
				}
			}
		}
	}
	return sortedKeys(dirty)
}
