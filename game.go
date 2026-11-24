package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type snowball struct {
	Sprite   *shape
	Location vec2
	Exists   bool
}

type levelone struct {
	ForegroundSprite *shape
	BackgroundSprite *shape
	Exists           bool
}

type levelzero struct {
	WelcomeText pointstext
	Exists      bool
}

type leveltwo struct {
	EndText pointstext
	Exists  bool
}

type snowman struct {
	Sprite *shape
	Up     bool
	Exists bool
}

type pointstext struct {
	Text   *label
	Exists bool
}

type music struct {
	Explosions []*sound
	Throws     []*sound
}

type textures struct {
	Foreground *ebiten.Image
	Background *ebiten.Image
	Snowman    *ebiten.Image
	Snowball   *ebiten.Image
}

type game struct {
	Gear          int
	LevelZero     levelzero
	LevelTwo      leveltwo
	Snowball      snowball
	Level         levelone
	Snowman       snowman
	SnowmanBall   snowball
	Time          int64
	Speed         int64
	Points        int
	SnowmanPoints int
	Font          *text.GoTextFace
	PointsText    pointstext
	Music         music
	Textures      textures
}

func NewGame() (*game, error) {
	game := new(game)
	game.Gear = 0

	font, err := loadFace("deja.ttf", 24)
	if err != nil {
		return nil, fmt.Errorf("loading font: %w", err)
	}
	game.Font = font

	for name, dst := range map[string]**ebiten.Image{
		"first-level-foreground.png": &game.Textures.Foreground,
		"first-level-background.png": &game.Textures.Background,
		"snowman.png":                &game.Textures.Snowman,
		"snowball.png":               &game.Textures.Snowball,
	} {
		img, err := loadImage(name)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", name, err)
		}
		*dst = img
	}

	for i := 1; i <= 4; i++ {
		explosion, err := loadSound("explosion" + strconv.Itoa(i) + ".wav")
		if err != nil {
			return nil, err
		}
		game.Music.Explosions = append(game.Music.Explosions, explosion)
		throw, err := loadSound("throw" + strconv.Itoa(i) + ".wav")
		if err != nil {
			return nil, err
		}
		game.Music.Throws = append(game.Music.Throws, throw)
	}

	rand.Seed(time.Now().UnixNano())
	game.InitGear()
	return game, nil
}

func (this *game) newText(s string, pos vec2) *label {
	return &label{Face: this.Font, String: s, Position: pos}
}

func (this *game) ChangeGearUp() {
	this.Gear += 1
	this.InitGear()
}

func (this *game) ChangeGearDown() {
	this.Gear -= 1
	this.InitGear()
}

func (this *game) InitGear() {
	if this.Gear == 0 { // the menu gear.
		var l levelzero
		l.WelcomeText.Text = this.newText("Welcome to the Snowball fights! Your duty is to win snowmen.\n\nYou can throw snowballs with your mouse!\nRemember to hit the snowmen or they will throw snowballs at you.\n\n\nPress enter to start the game.", vec2{300, 300})
		l.WelcomeText.Exists = true
		l.Exists = true
		this.LevelZero = l
	} else if this.Gear == 1 { // the playable gear.
		var l levelone
		l.ForegroundSprite = newShape(this.Textures.Foreground, vec2{1920, 1080}, vec2{0, 0})
		l.BackgroundSprite = newShape(this.Textures.Background, vec2{1920, 1080}, vec2{0, 0})
		l.Exists = true
		this.Level = l

		var s snowman
		s.Sprite = newShape(this.Textures.Snowman, vec2{184.0, 237.5}, vec2{1307.12, 624.84 + 50})
		s.Up = false
		s.Exists = true
		this.Snowman = s
		this.Time = time.Now().UnixNano()
		this.Points = 0
		this.SnowmanPoints = 0
		this.Speed = 4000000000
	} else if this.Gear == 2 {
		var l leveltwo
		var msg string
		if this.Points > this.SnowmanPoints {
			msg = "You have fallen. You made " + strconv.Itoa(this.Points) + " points\nwhile the snowman made only " + strconv.Itoa(this.SnowmanPoints) + " points.\n\nYou won the game!\n\nPress Enter if you want to play again.\nPress Escape to close the game."
		} else {
			msg = "You have fallen. You made only " + strconv.Itoa(this.Points) + " points\nwhile the snowman made " + strconv.Itoa(this.SnowmanPoints) + " points.\n\nYou lost the game!\n\nPress Enter if you want to play again.\nPress Escape to close the game."
		}
		l.EndText.Text = this.newText(msg, vec2{300, 300})
		l.EndText.Exists = true
		l.Exists = true
		this.LevelTwo = l
	}
}

func mousePosition() vec2 {
	x, y := ebiten.CursorPosition()
	return vec2{float64(x), float64(y)}
}

func (this *game) Update() error {
	if inpututil.IsKeyJustReleased(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyEnter) {
		if this.Gear == 0 {
			this.ChangeGearUp()
		} else if this.Gear == 2 {
			this.ChangeGearDown()
		}
	}
	if this.Gear == 1 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if this.Snowball.Exists == false {
				play := 1 + rand.Intn(4-1)
				this.Music.Throws[play].Play()
				var sb snowball
				sb.Location = mousePosition()
				sb.Sprite = newShape(this.Textures.Snowball, vec2{600, 600}, vec2{sb.Location.X - 300, sb.Location.Y - 300})
				sb.Exists = true
				this.Snowball = sb
			}
		}
		if this.Snowball.Exists == true {
			size := this.Snowball.Sprite.Size
			newSize := vec2{size.X - size.X*0.075, size.Y - size.Y*0.075}
			this.Snowball.Sprite.Size = newSize
			this.Snowball.Sprite.Origin = vec2{newSize.X / 2, newSize.Y / 2}
			this.Snowball.Sprite.Rotate(30.0)
			this.Snowball.Sprite.Position = this.Snowball.Location
			if size.X < 5 && size.Y < 5 {
				this.Snowball.Exists = false
			}
			if size.X < 50 && size.Y < 50 {
				sbRect := this.Snowball.Sprite.GlobalBounds()
				smRect := this.Snowman.Sprite.GlobalBounds()
				if sbRect.Intersects(smRect) {
					play := 1 + rand.Intn(4-1)
					this.Music.Explosions[play].Play()
					sSize := this.Snowman.Sprite.Size
					sPos := this.Snowman.Sprite.Position
					this.Snowman.Sprite.Move(vec2{0, sSize.Y})
					this.Snowman.Up = false
					// add points to the player!
					points := 500 + rand.Intn(1000-500)
					this.Points += points
					this.Speed -= int64(float64(this.Speed) * 0.05)

					this.Snowball.Exists = false
					this.PointsText.Text = this.newText("+"+strconv.Itoa(points), sPos)
					this.PointsText.Exists = true
				}
			}
		}
		if this.SnowmanBall.Exists == true {
			size := this.SnowmanBall.Sprite.Size
			newSize := vec2{size.X + size.X*0.075, size.Y + size.Y*0.075}
			this.SnowmanBall.Sprite.Size = newSize
			this.SnowmanBall.Sprite.Origin = vec2{newSize.X / 2, newSize.Y / 2}
			this.SnowmanBall.Sprite.Rotate(30.0)
			if newSize.X > 1200 && newSize.Y > 1200 {
				play := 1 + rand.Intn(4-1)
				this.Music.Explosions[play].Play()
				this.SnowmanBall.Exists = false
				// Add snowman hits at you.
				this.SnowmanPoints += 500 + rand.Intn(1000-500)
				if this.SnowmanPoints > 5500 {
					// The snowman won the game.
					this.ChangeGearUp()
					this.Snowman.Exists = false
				}
			}
		}
		if this.PointsText.Exists == true {
			old := this.PointsText.Text.Position
			if old.Y > -100 {
				this.PointsText.Text.Position = vec2{old.X, old.Y - 10}
			} else {
				this.PointsText.Exists = false
			}
		}
		if (this.Time + this.Speed) < time.Now().UnixNano() {
			if this.Snowman.Up == true {
				play := 1 + rand.Intn(4-1)
				this.Music.Throws[play].Play()
				// The snowman throws a snowball and drops down.
				var sb snowball
				sb.Location = mousePosition()
				sb.Sprite = newShape(this.Textures.Snowball, vec2{50, 50}, this.Snowman.Sprite.Position)
				sb.Exists = true
				this.SnowmanBall = sb
				sSize := this.Snowman.Sprite.Size
				this.Snowman.Sprite.Move(vec2{0, sSize.Y})
				this.Snowman.Up = false
			} else {
				// Change the snowman position.
				random := 1 + rand.Intn(8-1)
				switch random {
				case 1:
					this.Snowman.Sprite.Position = vec2{112.03, 711.39 + 50}
				case 2:
					this.Snowman.Sprite.Position = vec2{480.16, 757.38 + 50}
				case 3:
					this.Snowman.Sprite.Position = vec2{654.89, 806.07 + 50}
				case 4:
					this.Snowman.Sprite.Position = vec2{886.97, 765.49 + 50}
				case 5:
					this.Snowman.Sprite.Position = vec2{1099.04, 684.34 + 50}
				case 6:
					this.Snowman.Sprite.Position = vec2{1307.12, 624.84 + 50}
				case 7:
					this.Snowman.Sprite.Position = vec2{1399.15, 627.54 + 50}
				case 8:
					this.Snowman.Sprite.Position = vec2{1588.55, 691.11 + 50}
				}
				// Raise the snowman
				sSize := this.Snowman.Sprite.Size
				this.Snowman.Sprite.Move(vec2{0, -sSize.Y})
				this.Snowman.Up = true
			}
			this.Time = time.Now().UnixNano()
		}
	}
	return nil
}

func (this *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	if this.Gear == 0 {
		if this.LevelZero.Exists == true && this.LevelZero.WelcomeText.Exists == true {
			this.LevelZero.WelcomeText.Text.Draw(screen)
		}
	} else if this.Gear == 1 {
		if this.Level.Exists == true {
			this.Level.BackgroundSprite.Draw(screen)
		}
		if this.Snowman.Exists == true {
			this.Snowman.Sprite.Draw(screen)
		}
		if this.Level.Exists == true {
			this.Level.ForegroundSprite.Draw(screen)
		}
		if this.Snowball.Exists == true {
			this.Snowball.Sprite.Draw(screen)
		}
		if this.SnowmanBall.Exists == true {
			this.SnowmanBall.Sprite.Draw(screen)
		}
		if this.PointsText.Exists == true {
			this.PointsText.Text.Draw(screen)
		}
	} else if this.Gear == 2 {
		if this.LevelTwo.Exists == true && this.LevelTwo.EndText.Exists == true {
			this.LevelTwo.EndText.Text.Draw(screen)
		}
	}
}

func (this *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
