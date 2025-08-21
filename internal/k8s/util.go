/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package k8s

import (
	"fmt"
	"time"
)

func HumanReadableAge(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	d := time.Since(t)
	if d.Hours() > 24*365 {
		return fmt.Sprintf("%.0fy", d.Hours()/(24*365))
	}
	if d.Hours() > 24*30 {
		return fmt.Sprintf("%.0fmo", d.Hours()/(24*30))
	}
	if d.Hours() > 24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
