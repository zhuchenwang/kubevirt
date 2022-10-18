package rest

import (
	"fmt"

	restful "github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/client-go/log"

	apimetrics "kubevirt.io/kubevirt/pkg/monitoring/api"
	"kubevirt.io/kubevirt/pkg/util"
)

func (app *SubresourceAPIApp) VSOCKRequestHandler(request *restful.Request, response *restful.Response) {
	activeConnectionMetric := apimetrics.NewActiveVNCConnection(request.PathParameter("namespace"), request.PathParameter("name"))
	defer activeConnectionMetric.Dec()

	streamer := NewRawStreamer(
		app.FetchVirtualMachineInstance,
		validateVMIForVSOCK,
		app.virtHandlerDialer(func(vmi *v1.VirtualMachineInstance, conn kubecli.VirtHandlerConn) (string, error) {
			return conn.VSOCKURI(vmi, request.QueryParameter("port"))
		}),
	)

	streamer.Handle(request, response)
}

func validateVMIForVSOCK(vmi *v1.VirtualMachineInstance) *errors.StatusError {
	if !util.IsAutoAttachVSOCK(vmi) {
		err := fmt.Errorf("VSOCK is not attached.")
		log.Log.Object(vmi).Reason(err).Error("Can't establish Vsock connection.")
		return errors.NewBadRequest(err.Error())
	}
	return nil
}
