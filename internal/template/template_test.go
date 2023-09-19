package template

import (
	"net/url"
	"testing"

	"github.com/tf2d2/tf2d2/internal/utils"

	"github.com/stretchr/testify/assert"
	"oss.terrastruct.com/d2/d2target"
)

const iconURL = "https://foo.bar"

func TestRender(t *testing.T) {
	icon, _ := url.Parse(iconURL)

	testCases := []struct {
		name    string
		shapes  []*d2target.Shape
		conns   []*d2target.Connection
		isError bool
	}{
		{
			name: "template render shapes with icons and connections",
			shapes: []*d2target.Shape{
				{
					ID:   "s1_id",
					Type: "s1_type",
					Text: d2target.Text{
						Label: "s1_label",
					},
					Icon: icon,
				},
				{
					ID:   "s2_id",
					Type: "s2_type",
					Text: d2target.Text{
						Label: "s2_label",
					},
					Icon: icon,
				},
			},
			conns: []*d2target.Connection{
				{
					Src:      "s1_id",
					Dst:      "s2_id",
					DstArrow: d2target.ArrowArrowhead,
				},
			},
			isError: false,
		},
		{
			name: "template render shape with nil icon url",
			shapes: []*d2target.Shape{
				{
					ID:   "s1_id",
					Type: "s1_type",
					Text: d2target.Text{
						Label: "s1_label",
					},
					Icon: nil,
				},
			},
			conns:   nil,
			isError: true,
		},
		{
			name:    "template render no shapes",
			shapes:  nil,
			conns:   nil,
			isError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			expected, err := utils.GetExpected("shapes_icons_conns.expected")
			assert.Nil(err)

			tpl := New(tc.shapes, tc.conns)
			rendered, err := tpl.Render("")
			if tc.isError {
				assert.NotNil(err)
			} else {
				assert.Nil(err)
				assert.Equal(expected, rendered)
			}
		})
	}
}
