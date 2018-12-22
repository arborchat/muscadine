package tui

import "github.com/whereswaldon/gocui"

// BottomPrimaryLayout manages two gocui.Views within a gocui.Gui by respecting the size
// requests of the bottom view at the expense of the size of the top. The bottom view will
// be positioned at the bottom of the Gui, regardless of whatever internal position it may
// have requested. The bottom view will also be prevented from exapanding beyond half of the
// available screen real-estate.
//
// Note that BottomPrimaryLayout returns a Layout function, so proper usage is:
// ```
// manager := gocui.ManagerFunc(BottomPrimaryLayout("bottom-view-name", "top-view-name"))
// ```
func BottomPrimaryLayout(topView, bottomView string) func(*gocui.Gui) error {
	return func(g *gocui.Gui) error {
		// ensure both of our target view exist
		bottom, err := g.View(bottomView)
		if err != nil {
			return err
		}
		_, err = g.View(topView)
		if err != nil {
			return err
		}

		// check the size of the top view and make sure it's reasonable
		maxX, maxY := g.Size()
		_, bottomHeight := bottom.Size()
		if bottomHeight > maxY/2 {
			bottomHeight = maxY / 2
		}

		// configure the position and size of the bottom view
		bottomPosY := maxY - bottomHeight - 2
		if _, err = g.SetView(bottomView, 0, bottomPosY, maxX-1, maxY-1); err != nil {
			return err
		}

		// configure the position and size of the top view
		topHeight := bottomPosY - 1
		if _, err = g.SetView(topView, 0, 0, maxX-1, topHeight); err != nil {
			return err
		}
		return nil
	}
}
