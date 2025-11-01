package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	tb "gopkg.in/telebot.v3"
)

func main() {
	// Tokenni o‘zingiznikiga almashtiring
	const BOT_TOKEN = "8451386937:AAFatnFPs42izFlwiGjJip8Lb2crggA0jIk"

	pref := tb.Settings{
		Token:  BOT_TOKEN,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	// Ruxsat berilgan chatlar (kanal yoki guruhlar)
	allowedChats := map[int64]bool{
		-1003056945596: true,
	}

	// Reklama aniqlash regex
	linkRe := regexp.MustCompile(`(?i)(https?://|t\.me/|telegram\.me/|www\.)`)
	atRe := regexp.MustCompile(`(?m)@[\w\d_]{5,}`)

	blacklist := []string{"reklama", "advert", "sotiladi", "promo", "xiaomi", "telefon sotiladi", "shop"}

	// Xabar kelganda ishlaydigan handler
	bot.Handle(tb.OnText, func(c tb.Context) error {
		m := c.Message()
		if m == nil {
			return nil
		}

		chatID := m.Chat.ID
		if !allowedChats[chatID] {
			// Guruh ro‘yxatda bo‘lmasa, e’tibor bermaymiz
			return nil
		}

		text := m.Text
		if text == "" {
			text = m.Caption
		}

		// Reklama aniqlash
		isAd := false
		if linkRe.MatchString(text) || atRe.MatchString(text) {
			isAd = true
		} else {
			lower := strings.ToLower(text)
			for _, w := range blacklist {
				if strings.Contains(lower, w) {
					isAd = true
					break
				}
			}
		}

		if isAd {
			member, err := bot.ChatMemberOf(c.Chat(), m.Sender)
			if err != nil {
				log.Printf("ChatMember aniqlanmadi: %v", err)
				return nil
			}

			status := member.Role
			if status == tb.Creator || status == tb.Administrator {
				log.Printf("Admin xabari: %s", m.Sender.Username)
				return nil
			}

			if err := bot.Delete(m); err != nil {
				log.Printf("Xabarni o‘chirishda xatolik: %v", err)
			} else {
				log.Printf("Reklama o‘chirildi: %s", text)
			}

			// Foydalanuvchini ogohlantirish (faqat username ishlatamiz)
			var name string
			if m.Sender.Username != "" {
				name = "@" + m.Sender.Username
			} else {
				name = htmlEscape(m.Sender.FirstName + " " + m.Sender.LastName)
			}

			warn := fmt.Sprintf(`
			⚠️ %s!

			<b>Diqqat!</b> Bu guruhda reklama, havola yoki sotuvga oid xabarlar yuborish taqiqlangan.

			❌ Reklama, link, mention, yoki mahsulot/sotuv so‘zlari aniqlansa — xabar o‘chiriladi.
			🔒 Takror holatda foydalanuvchi vaqtincha bloklanishi mumkin.

				Iltimos, guruh qoidalariga amal qiling va tartibni saqlang. 🙏
			`, name)
			return c.Send(warn, tb.ModeHTML)
		}

		return nil
	})

	log.Println("🤖 Bot ishga tushdi...")
	bot.Start()
}

// HTML belgilarni escape qilish
func htmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(s)
}
