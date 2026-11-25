package labelling

import (
	"github.com/jastribl/photosync/photos"
)

// LabelPlacement is one text label to insert into an album, with an ordered
// list of anchor candidates. Anchors[0] is the preferred anchor; if adding
// the enrichment after it fails (for example that photo is not in the
// album), the caller tries Anchors[1], then Anchors[2], and so on. An empty
// Anchors list means the label's placement IS the front of the album
// (FIRST_IN_ALBUM) — that is the case for the last folder, which has no
// later folder to anchor after. The front is never used as a fallback: if
// Anchors is non-empty and every anchor fails, the label is not added at
// all and the failure is reported instead.
type LabelPlacement struct {
	Text    string
	Anchors []*photos.MediaItem
}

// BuildLabelPlan computes the full list of label placements from the
// per-folder info before any label is written, so the plan is a snapshot:
// a failure mid-run cannot leave the remaining labels unplanned.
//
// Anchor chain: a label's anchors are the later folders' photos,
// oldest first within each folder, so if the ideal anchor (the next
// folder's oldest photo) is not in the album, the label lands after the
// closest surviving photo to its correct spot instead of jumping a whole
// day. Each folder contributes at most maxAnchorsPerFolder candidates: if
// that many of a folder's oldest photos are all missing from the album,
// the folder is effectively gone, and the chain moves on to the next
// folder instead of burning API calls on every photo in it.
//
// Ordering assumption: the album is sorted newest first (which is how
// Justin keeps his albums). A folder's label goes right after the last
// photo of the next folder that has photos, and the last folder's label
// (or any label with no later folder) goes at the front of the album — the
// front is the newest end, so that is where the newest folder's label
// belongs. The anchor-relative placements survive a re-sort (each label
// stays attached to its anchor photo), but the front placement does not:
// in an oldest-first album the newest folder's label would land at the
// wrong end. Folder order comes from the local directory listing, which
// is alphabetical by folder name, so it is chronological only when folder
// names sort chronologically (e.g. ISO dates like 2026-11-20).
//
// maxAnchorsPerFolder bounds how many photos of one later folder a label
// will try as anchors. See the anchor-chain note above.
const maxAnchorsPerFolder = 10

// anchorCandidates returns the folder's picked photos, oldest first,
// capped at maxAnchorsPerFolder.
func anchorCandidates(folder *FolderInfo) []*photos.MediaItem {
	if len(folder.MediaItems) > maxAnchorsPerFolder {
		return folder.MediaItems[:maxAnchorsPerFolder]
	}
	return folder.MediaItems
}
func BuildLabelPlan(listOfFolderInfo []*FolderInfo, rootDir string) []LabelPlacement {
	plan := []LabelPlacement{}
	for i, folderInfo := range listOfFolderInfo {
		if folderInfo.NumMediaItems == 0 {
			continue
		}
		if folderInfo.Path == rootDir {
			continue
		}
		labelText := folderInfo.Path[len(rootDir) : len(folderInfo.Path)-1]
		placement := LabelPlacement{Text: labelText}
		if i != len(listOfFolderInfo)-1 {
			for j := i + 1; j < len(listOfFolderInfo); j++ {
				placement.Anchors = append(placement.Anchors, anchorCandidates(listOfFolderInfo[j])...)
			}
		}
		plan = append(plan, placement)
	}
	return plan
}
