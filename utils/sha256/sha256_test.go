package sha256

import (
	"fmt"
	"testing"

	"github.com/skirrund/gcloud/utils"
)

type Ttt struct {
	Date *utils.DateTime
}

func TestSha256(t *testing.T) {
	str := "mobile%3D13321954022%26path_type%3D1%26secret_id%3Dqm07113040202069900%26timestamp%3D1693381409%26secret_key%3Db7a3d9bd91ed38edd91e980afd747dd8"
	fmt.Println(Encry(str))
	fmt.Println(Sha256Hex([]byte(str)))
}
