/* package main

import (
	"fmt"










	"github.com/hpcloud/tail"
)

func main() {
	t, err := tail.TailFile("/tmp/foo", tail.Config{Follow: true})
	if err != nil {
		panic(err)
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}
*/

package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"os"
	"time"
)

func main() {
	//filename := "/var/lib/docker/containers/02fcc5b379556940b5b33037ce55adc23fc0c8f85c21574dcd68c86c516c9547/02fcc5b379556940b5b33037ce55adc23fc0c8f85c21574dcd68c86c516c9547-json.log"
	filename := "/tmp/foo"
	tails, err := tail.TailFile(filename, tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}, //tail from the last Nth location
		MustExist: false,
		Poll:      true,
	})
	if err != nil {
		fmt.Println("tail file err:", err)
		return
	}
	var msg *tail.Line
	var ok bool
	for true {
		msg, ok = <-tails.Lines
		if !ok {
			fmt.Printf("tail file close reopen, filename:%s\n", tails.Filename)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		fmt.Println("msg:", msg)
		fmt.Println("msg text:", msg.Text)
	}
}
