package main

import (
    "fmt"
    "os/exec"
    "strings"
    "golang.org/x/text/unicode/norm"

    "github.com/gdamore/tcell/v2"
    "github.com/mattn/go-runewidth"
)

func main() {
    screen, err := tcell.NewScreen()
    if err != nil {
        panic(err)
    }
    if err := screen.Init(); err != nil {
        panic(err)
    }
    defer screen.Fini()

    screen.Clear()

    text1 := ""
    text2 := ""
    result := ""

    activePanel := 0 // 0 = левая, 1 = правая, 2 = результат
    activeButton := 0

    // Смещения для скролла
    text1Offset := 0
    text2Offset := 0
    resultOffset := 0

    // Флаг для отображения окна справки
    showHelp := false

    draw := func() {
        screen.Clear()
        w, h := screen.Size()

        upperHeight := (h * 2) / 3
        lowerHeight := h - upperHeight

        // Левая панель
        drawBox(screen, 0, 0, w/2, upperHeight, "Текст 1", activePanel == 0)
        drawText(screen, 1, 1, text1, w/2-2, upperHeight-2, text1Offset)

        // Правая панель
        drawBox(screen, w/2, 0, w/2, upperHeight, "Текст 2", activePanel == 1)
        drawText(screen, w/2+1, 1, text2, w/2-2, upperHeight-2, text2Offset)

        // Кнопки
        btnStyle := func(active bool) tcell.Style {
            if active {
                return tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
            }
            return tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
        }
        printText(screen, 2, upperHeight, " Сравнить ", btnStyle(activeButton == 0))
        printText(screen, 20, upperHeight, " Очистить ", btnStyle(activeButton == 1))

        // Результат
        drawBox(screen, 0, upperHeight+1, w, lowerHeight-1, "Результат", activePanel == 2)
        drawScrollableText(screen, 1, upperHeight+2, result, w-2, lowerHeight-3, resultOffset)

        // Отображение окна справки, если установлен флаг
        if showHelp {
            drawHelpPopup(screen, w, h)
        }

        screen.Show()
    }

    for {
        ev := screen.PollEvent()
        switch ev := ev.(type) {
        case *tcell.EventKey:
            switch ev.Key() {
            case tcell.KeyCtrlC:
                return

            case tcell.KeyRune:
                if ev.Rune() == '?' {
                    // Переключаем флаг отображения справки
                    showHelp = !showHelp

                } else if ev.Rune() != 0 {
                    if activePanel == 0 {
                        text1 += string(ev.Rune())
                    } else if activePanel == 1 {
                        text2 += string(ev.Rune())
                    }
                }

            case tcell.KeyTab:
                activePanel = (activePanel + 1) % 3

            case tcell.KeyRight:
                activeButton = (activeButton + 1) % 2

            case tcell.KeyLeft:
                activeButton = (activeButton + 1) % 2

            case tcell.KeyEnter:
                if activeButton == 0 {
                    result = compareTexts(text1, text2)
                    resultOffset = 0
                } else if activeButton == 1 {
                    if activePanel == 0 {
                        text1 = ""
                        text1Offset = 0
                    } else if activePanel == 1 {
                        text2 = ""
                        text2Offset = 0
                    } else if activePanel == 2 {
                        result = ""
                        resultOffset = 0
                    }
                }

            case tcell.KeyCtrlV:
                buf := readClipboard()
                if activePanel == 0 {
                    text1 += buf
                } else if activePanel == 1 {
                    text2 += buf
                }

            case tcell.KeyBackspace, tcell.KeyBackspace2:
                if activePanel == 0 && len(text1) > 0 {
                    text1 = text1[:len(text1)-1]
                } else if activePanel == 1 && len(text2) > 0 {
                    text2 = text2[:len(text2)-1]
                }

            case tcell.KeyUp:
                if activePanel == 0 && text1Offset > 0 {
                    text1Offset--
                } else if activePanel == 1 && text2Offset > 0 {
                    text2Offset--
                } else if activePanel == 2 && resultOffset > 0 {
                    resultOffset--
                }

            case tcell.KeyDown:
                w, h := screen.Size()
                upperHeight := (h * 2) / 3
                lowerHeight := h - upperHeight

                if activePanel == 0 {
                    maxOffset := calculateMaxOffset(text1, w/2-2, upperHeight-2)
                    if text1Offset < maxOffset {
                        text1Offset++
                    }
                } else if activePanel == 1 {
                    maxOffset := calculateMaxOffset(text2, w/2-2, upperHeight-2)
                    if text2Offset < maxOffset {
                        text2Offset++
                    }
                } else if activePanel == 2 {
                    maxOffset := calculateMaxOffset(result, w-2, lowerHeight-3)
                    if resultOffset < maxOffset {
                        resultOffset++
                    }
                }

            }
            draw()

        case *tcell.EventResize:
            screen.Sync()
            draw()
        }
    }
}

func drawBox(s tcell.Screen, x, y, w, h int, title string, active bool) {
    var borderStyle tcell.Style
    if active {
        borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDefault)
    } else {
        borderStyle = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDefault)
    }

    fillStyle := tcell.StyleDefault.Background(tcell.ColorDefault)

    for i := x; i < x+w; i++ {
        for j := y; j < y+h; j++ {
            s.SetContent(i, j, ' ', nil, fillStyle)
        }
    }

    for i := x; i < x+w; i++ {
        s.SetContent(i, y, tcell.RuneHLine, nil, borderStyle)
        s.SetContent(i, y+h-1, tcell.RuneHLine, nil, borderStyle)
    }
    for j := y; j < y+h; j++ {
        s.SetContent(x, j, tcell.RuneVLine, nil, borderStyle)
        s.SetContent(x+w-1, j, tcell.RuneVLine, nil, borderStyle)
    }

    s.SetContent(x, y, tcell.RuneULCorner, nil, borderStyle)
    s.SetContent(x+w-1, y, tcell.RuneURCorner, nil, borderStyle)
    s.SetContent(x, y+h-1, tcell.RuneLLCorner, nil, borderStyle)
    s.SetContent(x+w-1, y+h-1, tcell.RuneLRCorner, nil, borderStyle)

    printText(s, x+2, y, " "+title+" ", borderStyle)
}

// drawText с нумерацией строк и скроллом
func drawText(s tcell.Screen, x, y int, text string, maxWidth, maxHeight, offset int) {
    style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
    lines := strings.Split(text, "\n")

    startIndex := offset
    endIndex := offset + maxHeight
    if endIndex > len(lines) {
        endIndex = len(lines)
    }

    for i, line := range lines[startIndex:endIndex] {
        lineNum := fmt.Sprintf("%3d   ", i+1+offset)
        fullLine := lineNum + truncate(line, maxWidth-len(lineNum))
        printText(s, x, y+i, fullLine, style)
    }
}

// drawScrollableText для результата
func drawScrollableText(s tcell.Screen, x, y int, text string, maxWidth, maxHeight int, offset int) {
    lines := strings.Split(text, "\n")
    startIndex := offset
    endIndex := offset + maxHeight

    if startIndex > len(lines) {
        return
    }
    if endIndex > len(lines) {
        endIndex = len(lines)
    }

    for i, line := range lines[startIndex:endIndex] {
        pos := x

        // Добавляем номер строки
        lineNumber := fmt.Sprintf("%3d   ", startIndex+i+1)
        printText(s, pos, y+i, lineNumber, tcell.StyleDefault.Foreground(tcell.ColorGray))
        pos += runewidth.StringWidth(lineNumber)

        // разделяем по "≠"
        parts := strings.SplitN(line, "≠", 2)

        // левая часть (серым)
        if len(parts) > 0 && parts[0] != "" {
            printText(s, pos, y+i, truncate(parts[0], maxWidth-pos), tcell.StyleDefault.Foreground(tcell.ColorGray))
            pos += runewidth.StringWidth(parts[0])
        }

        // если есть правая часть — добавляем символ "≠" и правый текст
        if len(parts) == 2 {
            printText(s, pos, y+i, " ≠ "+truncate(parts[1], maxWidth-pos), tcell.StyleDefault.Foreground(tcell.ColorWhite))
        }
    }
}

func calculateMaxOffset(text string, maxWidth, maxHeight int) int {
    lines := strings.Split(text, "\n")
    totalLines := len(lines)
    if totalLines <= maxHeight {
        return 0
    }
    return totalLines - maxHeight
}

func printText(s tcell.Screen, x, y int, text string, style tcell.Style) {
    pos := x
    for _, ch := range text {
        width := runewidth.RuneWidth(ch)
        s.SetContent(pos, y, ch, nil, style)
        pos += width
    }
}

func truncate(s string, max int) string {
    if len(s) > max {
        return s[:max]
    }
    return s
}

func normalizeForCompare(s string) string {
    // Заменяем NBSP на обычный пробел и удаляем некоторые zero-width символы
    s = strings.Map(func(r rune) rune {
        switch r {
        case '\u00A0':
            return ' '
        case '\u200B', '\u200C', '\u200D', '\uFEFF':
            return -1
        default:
            return r
        }
    }, s)

    // Unicode нормализация (NFC)
    s = norm.NFC.String(s)

    // Убираем только хвостовые пробелы/табы (не удаляем ведущие, чтобы не менять вид в UI)
    s = strings.TrimRight(s, " \t")

    return s
}

func compareTexts(a, b string) string {
    // Нормализуем окончания строк
    a = strings.ReplaceAll(a, "\r\n", "\n")
    b = strings.ReplaceAll(b, "\r\n", "\n")

    // Быстрая проверка: если тексты полностью совпадают после нормализации, вернуть короткое сообщение
    if norm.NFC.String(a) == norm.NFC.String(b) {
        return "Тексты идентичны"
    }

    linesA := strings.Split(a, "\n")
    linesB := strings.Split(b, "\n")
    maxLines := len(linesA)
    if len(linesB) > maxLines {
        maxLines = len(linesB)
    }

    var sb strings.Builder
    for i := 0; i < maxLines; i++ {
        var origA, origB string
        if i < len(linesA) {
            origA = linesA[i]
        }
        if i < len(linesB) {
            origB = linesB[i]
        }

        cmpA := normalizeForCompare(origA)
        cmpB := normalizeForCompare(origB)

        if cmpA == cmpB {
            // строки равны после нормализации — помечаем ≡ (или можно использовать обычный разделитель)
            sb.WriteString(origA)
            sb.WriteString(" ≡ ")
            sb.WriteString(origB)
            sb.WriteByte('\n')
        } else {
            // различаются — ставим ≠
            sb.WriteString(origA)
            sb.WriteString(" ≠ ")
            sb.WriteString(origB)
            sb.WriteByte('\n')
        }
    }

    return sb.String()
}
func readClipboard() string {
    out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
    if err != nil {
        fmt.Println("Error reading clipboard:", err)
        return ""
    }
    return string(out)
}

func drawHelpPopup(s tcell.Screen, screenWidth, screenHeight int) {
    // Размеры окна справки
    width := screenWidth / 2
    height := screenHeight / 2

    // Позиция окна справки (в центре экрана)
    x := (screenWidth - width) / 2
    y := (screenHeight - height) / 2

    // Стиль рамки, текста и фона
    borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorDefault)
    textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
    backgroundStyle := tcell.StyleDefault.Background(tcell.ColorDefault) // Черный фон

    // Рисуем фон
    for i := x; i < x+width; i++ {
        for j := y; j < y+height; j++ {
            s.SetContent(i, j, ' ', nil, backgroundStyle)
        }
    }

    // Рисуем рамку
    for i := x; i < x+width; i++ {
        s.SetContent(i, y, tcell.RuneHLine, nil, borderStyle)
        s.SetContent(i, y+height-1, tcell.RuneHLine, nil, borderStyle)
    }
    for j := y; j < y+height; j++ {
        s.SetContent(x, j, tcell.RuneVLine, nil, borderStyle)
        s.SetContent(x+width-1, j, tcell.RuneVLine, nil, borderStyle)
    }
    s.SetContent(x, y, tcell.RuneULCorner, nil, borderStyle)
    s.SetContent(x+width-1, y, tcell.RuneURCorner, nil, borderStyle)
    s.SetContent(x, y+height-1, tcell.RuneLLCorner, nil, borderStyle)
    s.SetContent(x+width-1, y+height-1, tcell.RuneLRCorner, nil, borderStyle)

    // Заголовок
    printText(s, x+2, y, " Справка (Горячие клавиши) ", borderStyle)

    // Текст справки
    helpText := []string{
        "Ctrl+C       Выход",
        "Tab          Переключение между панелями",
        "Ctrl+V       Вставить из буфера обмена",
        "Backspace    Удалить символ",
        "Up/Down      Скролл текста",
        "Enter        Выполнить действие кнопки",
        "Left/Right   Смена активной кнопки",
        "?            Показать/скрыть справку",
    }

    // Вывод текста справки
    for i, line := range helpText {
        printText(s, x+2, y+2+i, line, textStyle)
    }
}
