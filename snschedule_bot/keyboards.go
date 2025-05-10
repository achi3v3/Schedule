package functions

import (
	"github.com/go-telegram/bot/models"
)

type Builder struct {
	rows [][]models.InlineKeyboardButton
}

func New() *Builder {
	return &Builder{}
}

func (k *Builder) AddRow(buttons ...models.InlineKeyboardButton) *Builder {
	k.rows = append(k.rows, buttons)
	return k
}

func (k *Builder) AddButtonsInRow(buttons ...models.InlineKeyboardButton) *Builder {
	if len(k.rows) == 0 {
		k.rows = append(k.rows, []models.InlineKeyboardButton{})
	}

	lastRowIdx := len(k.rows) - 1
	k.rows[lastRowIdx] = append(k.rows[lastRowIdx], buttons...)
	return k
}

func (k *Builder) Build() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: k.rows,
	}
}
