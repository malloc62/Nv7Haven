package elements

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/Nv7-Github/sevcord/v2"
)

func (e *Elements) ImageCmd(c sevcord.Ctx, opts []any) {
	c.Acknowledge()

	// Check element
	var elem string
	var old string
	err := e.db.QueryRow("SELECT name, image FROM elements WHERE id=$1 AND guild=$2", opts[0].(int64), c.Guild()).Scan(&elem, &old)
	if err != nil {
		e.base.Error(c, err)
		return
	}

	// Check image
	if !strings.HasPrefix(opts[1].(*sevcord.SlashCommandAttachment).ContentType, "image") {
		c.Respond(sevcord.NewMessage("The attachment must be an image! " + types.RedCircle))
		return
	}

	// Make poll
	e.polls.CreatePoll(c, &types.Poll{
		Kind: types.PollKindImage,
		Data: types.PgData{
			"elem": float64(opts[0].(int64)),
			"new":  opts[1].(*sevcord.SlashCommandAttachment).URL,
			"old":  old,
		},
	})

	// Respond
	c.Respond(sevcord.NewMessage(fmt.Sprintf("Suggested an image for **%s** 📷", elem)))
}

func (e *Elements) SignCmd(c sevcord.Ctx, opts []any) {
	// Check element
	var name string
	var old string
	err := e.db.QueryRow("SELECT name, comment FROM elements WHERE id=$1 AND guild=$2", opts[0].(int64), c.Guild()).Scan(&name, &old)
	if err != nil {
		e.base.Error(c, err)
		return
	}

	// Get mark
	c.(*sevcord.InteractionCtx).Modal(sevcord.NewModal("Sign Element", func(c sevcord.Ctx, s []string) {
		// Make poll
		e.polls.CreatePoll(c, &types.Poll{
			Kind: types.PollKindComment,
			Data: types.PgData{
				"elem": float64(opts[0].(int64)),
				"new":  s[0],
				"old":  old,
			},
		})

		// Respond
		c.Respond(sevcord.NewMessage(fmt.Sprintf("Suggested a note for **%s** 🖋️", name)))
	}).Input(sevcord.NewModalInput("New Comment", "None", sevcord.ModalInputStyleParagraph, 2400)))
}

func (e *Elements) ColorCmd(c sevcord.Ctx, opts []any) {
	c.Acknowledge()

	// Check hex code
	code := opts[1].(string)
	if !strings.HasPrefix(code, "#") {
		c.Respond(sevcord.NewMessage("Invalid hex code! " + types.RedCircle))
		return
	}
	val, err := strconv.ParseInt(strings.TrimPrefix(code, "#"), 16, 64)
	if err != nil {
		e.base.Error(c, err)
		return
	}
	if val < 0 || val > 16777215 {
		c.Respond(sevcord.NewMessage("Invalid hex code! " + types.RedCircle))
		return
	}

	// Check element
	var name string
	var old int
	err = e.db.QueryRow("SELECT name, color FROM elements WHERE id=$1 AND guild=$2", opts[0].(int64), c.Guild()).Scan(&name, &old)
	if err != nil {
		e.base.Error(c, err)
		return
	}

	// Make poll
	e.polls.CreatePoll(c, &types.Poll{
		Kind: types.PollKindColor,
		Data: types.PgData{
			"elem": float64(opts[0].(int64)),
			"new":  float64(val),
			"old":  float64(old),
		},
	})

	// Respond
	c.Respond(sevcord.NewMessage(fmt.Sprintf("Suggested a color for **%s** 🎨", name)))
}
