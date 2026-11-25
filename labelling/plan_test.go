package labelling

import (
	"fmt"
	"testing"

	"github.com/jastribl/photosync/photos"
)

func item(id, filename string) *photos.MediaItem {
	return &photos.MediaItem{ID: id, Filename: filename}
}

func folder(path string, numItems int, items ...*photos.MediaItem) *FolderInfo {
	f := &FolderInfo{Path: path, NumMediaItems: numItems, MediaItems: items}
	if len(items) > 0 {
		f.LastMediaItem = items[0]
	}
	return f
}

func TestBuildLabelPlanBasic(t *testing.T) {
	// Three folders, each with photos: every label anchors after the next
	// folder's last photo, and the last folder's label goes to the front.
	folders := []*FolderInfo{
		folder("/pics/2024-01/", 2, item("a1", "a.jpg")),
		folder("/pics/2024-02/", 3, item("b1", "b.jpg")),
		folder("/pics/2024-03/", 1, item("c1", "c.jpg")),
	}
	plan := BuildLabelPlan(folders, "/pics/")
	if len(plan) != 3 {
		t.Fatalf("expected 3 placements, got %d", len(plan))
	}
	if plan[0].Text != "2024-01" || len(plan[0].Anchors) != 2 || plan[0].Anchors[0].ID != "b1" || plan[0].Anchors[1].ID != "c1" {
		t.Errorf("first label wrong: %+v", plan[0])
	}
	if plan[1].Text != "2024-02" || len(plan[1].Anchors) != 1 || plan[1].Anchors[0].ID != "c1" {
		t.Errorf("second label wrong: %+v", plan[1])
	}
	if plan[2].Text != "2024-03" || len(plan[2].Anchors) != 0 {
		t.Errorf("last label should go to the front (no anchors): %+v", plan[2])
	}
}

func TestBuildLabelPlanSkipsEmptyFolders(t *testing.T) {
	// The middle folder has no photos in the album: it gets no label, and
	// the first folder's anchor walks forward past it to the third folder.
	folders := []*FolderInfo{
		folder("/pics/2024-01/", 2, item("a1", "a.jpg")),
		folder("/pics/2024-02/", 0),
		folder("/pics/2024-03/", 1, item("c1", "c.jpg")),
	}
	plan := BuildLabelPlan(folders, "/pics/")
	if len(plan) != 2 {
		t.Fatalf("expected 2 placements, got %d", len(plan))
	}
	if plan[0].Text != "2024-01" || len(plan[0].Anchors) != 1 || plan[0].Anchors[0].ID != "c1" {
		t.Errorf("first label should anchor after the third folder's photo: %+v", plan[0])
	}
	if plan[1].Text != "2024-03" || len(plan[1].Anchors) != 0 {
		t.Errorf("last label should go to the front: %+v", plan[1])
	}
}

func TestBuildLabelPlanAllEmpty(t *testing.T) {
	folders := []*FolderInfo{
		folder("/pics/2024-01/", 0),
		folder("/pics/2024-02/", 0),
	}
	if plan := BuildLabelPlan(folders, "/pics/"); len(plan) != 0 {
		t.Errorf("expected no placements, got %+v", plan)
	}
}

func TestBuildLabelPlanSkipsRootDir(t *testing.T) {
	folders := []*FolderInfo{
		folder("/pics/", 2, item("a1", "a.jpg")),
		folder("/pics/2024-01/", 1, item("b1", "b.jpg")),
	}
	plan := BuildLabelPlan(folders, "/pics/")
	if len(plan) != 1 || plan[0].Text != "2024-01" {
		t.Errorf("root dir should get no label, got %+v", plan)
	}
}

func TestBuildLabelPlanMultiPhotoChain(t *testing.T) {
	// The middle folder has three photos: the first label's anchors are
	// all three (oldest first), then the last folder's photo.
	folders := []*FolderInfo{
		folder("/pics/2024-01/", 1, item("a1", "a.jpg")),
		folder("/pics/2024-02/", 3, item("b1", "b.jpg"), item("b2", "b2.jpg"), item("b3", "b3.jpg")),
		folder("/pics/2024-03/", 1, item("c1", "c.jpg")),
	}
	plan := BuildLabelPlan(folders, "/pics/")
	if len(plan) != 3 {
		t.Fatalf("expected 3 placements, got %d", len(plan))
	}
	want := []string{"b1", "b2", "b3", "c1"}
	if got := anchorIDs(plan[0]); !equalStrings(got, want) {
		t.Errorf("first label anchors = %v, want %v", got, want)
	}
	if got := anchorIDs(plan[1]); !equalStrings(got, []string{"c1"}) {
		t.Errorf("second label anchors = %v, want [c1]", got)
	}
	if len(plan[2].Anchors) != 0 {
		t.Errorf("last label should go to the front (no anchors): %+v", plan[2])
	}
}

func TestBuildLabelPlanCapsAnchorsPerFolder(t *testing.T) {
	// A folder with more than maxAnchorsPerFolder photos contributes
	// only the oldest maxAnchorsPerFolder to a label's chain.
	var middle []*photos.MediaItem
	for i := 1; i <= maxAnchorsPerFolder+2; i++ {
		middle = append(middle, item(fmt.Sprintf("b%d", i), fmt.Sprintf("b%d.jpg", i)))
	}
	folders := []*FolderInfo{
		folder("/pics/2024-01/", 1, item("a1", "a.jpg")),
		{Path: "/pics/2024-02/", NumMediaItems: len(middle), LastMediaItem: middle[0], MediaItems: middle},
		folder("/pics/2024-03/", 1, item("c1", "c.jpg")),
	}
	plan := BuildLabelPlan(folders, "/pics/")
	if len(plan) != 3 {
		t.Fatalf("expected 3 placements, got %d", len(plan))
	}
	want := []string{"b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8", "b9", "b10", "c1"}
	if got := anchorIDs(plan[0]); !equalStrings(got, want) {
		t.Errorf("first label anchors = %v, want %v", got, want)
	}
}

func anchorIDs(p LabelPlacement) []string {
	var ids []string
	for _, a := range p.Anchors {
		ids = append(ids, a.ID)
	}
	return ids
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
