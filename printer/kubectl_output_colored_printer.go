package printer

import (
	"io"
	"strconv"
	"strings"

	"github.com/hidetatz/kubecolor/color"
	"github.com/hidetatz/kubecolor/kubectl"
)

// KubectlOutputColoredPrinter is a printer to print data depending on
// which kubectl subcommand is executed.
type KubectlOutputColoredPrinter struct {
	SubcommandInfo *kubectl.SubcommandInfo
	DarkBackground bool
	Recursive      bool
}

// Print reads r then write it to w, its format is based on kubectl subcommand.
// If given subcommand is not supported by the printer, it prints data in Green.
func (kp *KubectlOutputColoredPrinter) Print(r io.Reader, w io.Writer) {
	withHeader := !kp.SubcommandInfo.NoHeader

	var printer Printer = &SingleColoredPrinter{Color: color.Green} // default in green

	switch kp.SubcommandInfo.Subcommand {
	case kubectl.Top, kubectl.APIResources:
		printer = NewTablePrinter(withHeader, kp.DarkBackground, nil)

	case kubectl.APIVersions:
		printer = NewTablePrinter(false, kp.DarkBackground, nil) // api-versions always doesn't have header

	case kubectl.Get:
		switch {
		case kp.SubcommandInfo.FormatOption == kubectl.None, kp.SubcommandInfo.FormatOption == kubectl.Wide:
			printer = NewTablePrinterWithHeader(
				withHeader,
				kp.DarkBackground,
				func(i int, column string, header string) (color.Color, bool) {
					if header == "STATUS" || (header == "" && i == 2) {
						switch column {
						case "Running":
							return color.Green256, true
						case "Completed", "Succeeded":
							return color.DarkGreen, true
						case "Pending", "ContainerCreating":
							return color.Yellow256, true
						case "ImagePullBackOff", "CrashLoopBackOff", "Error":
							return color.Red, true
						}
					}

					// When Readiness is "n/m" then yellow if n > 0, red if n == 0; Green if n == m
					if header == "READY" || (header == "" && strings.Count(column, "/") == 1) {
						if arr := strings.Split(column, "/"); arr[0] != arr[1] {
							n, e1 := strconv.Atoi(arr[0])
							m, e2 := strconv.Atoi(arr[1])
							if e1 == nil && e2 == nil { // check both is number
								if n == 0 && m > 0 {
									return color.Red, true
								}
								return color.Yellow256, true
							}
						} else {
							return color.Green256, true
						}
					}

					// Green if 0 restarts, yellow if > 0, red if can't parse
					if header == "RESTARTS" || (header == "" && i == 3) {
						if r, err := strconv.Atoi(column); err != nil {
							return color.Red, true
						} else {
							if r == 0 {
								return color.Green256, true
							}
							return color.Yellow256, true
						}
					}

					// More intense yellow the more recent it is, less saturated the older
					// E.g. Yellow256 if age is 26s, Yellow256 + 4 (dim yellow) if age is 1y56d
					if header == "AGE" || (header == "" && i == 4) {
						const units = "smhdy"
						offset := -1
						for i := len(units) - 1; i >= 0; i-- {
							if strings.Contains(column, string(units[i])) {
								offset = i
								break
							}
						}
						return color.Color(int(color.Yellow256) + offset), true
					}

					return 0, false
				},
			)
		case kp.SubcommandInfo.FormatOption == kubectl.Json:
			printer = &JsonPrinter{DarkBackground: kp.DarkBackground}
		case kp.SubcommandInfo.FormatOption == kubectl.Yaml:
			printer = &YamlPrinter{DarkBackground: kp.DarkBackground}
		}

	case kubectl.Describe:
		printer = &DescribePrinter{
			DarkBackground: kp.DarkBackground,
			TablePrinter:   NewTablePrinter(false, kp.DarkBackground, nil),
		}
	case kubectl.Explain:
		printer = &ExplainPrinter{
			DarkBackground: kp.DarkBackground,
			Recursive:      kp.Recursive,
		}
	case kubectl.Version:
		switch {
		case kp.SubcommandInfo.FormatOption == kubectl.Json:
			printer = &JsonPrinter{DarkBackground: kp.DarkBackground}
		case kp.SubcommandInfo.FormatOption == kubectl.Yaml:
			printer = &YamlPrinter{DarkBackground: kp.DarkBackground}
		case kp.SubcommandInfo.Short:
			printer = &VersionShortPrinter{
				DarkBackground: kp.DarkBackground,
			}
		default:
			printer = &VersionPrinter{
				DarkBackground: kp.DarkBackground,
			}
		}
	case kubectl.Options:
		printer = &OptionsPrinter{
			DarkBackground: kp.DarkBackground,
		}
	case kubectl.Apply:
		switch {
		case kp.SubcommandInfo.FormatOption == kubectl.Json:
			printer = &JsonPrinter{DarkBackground: kp.DarkBackground}
		case kp.SubcommandInfo.FormatOption == kubectl.Yaml:
			printer = &YamlPrinter{DarkBackground: kp.DarkBackground}
		default:
			printer = &ApplyPrinter{DarkBackground: kp.DarkBackground}
		}
	}

	if kp.SubcommandInfo.Help {
		printer = &SingleColoredPrinter{Color: color.Yellow}
	}

	printer.Print(r, w)
}
