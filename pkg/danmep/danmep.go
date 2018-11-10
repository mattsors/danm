package danmep

import (
  "log"
  "strconv"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  danmtypes "github.com/nokia/danm/pkg/crd/apis/danm/v1"
  danmclientset "github.com/nokia/danm/pkg/crd/client/clientset/versioned"
)

const (
  defaultNamespace = "default"
  dockerApiVersion = "1.22"
)

// DeleteIpvlanInterface deletes a Pod's IPVLAN network interface based on the related DanmEp
func DeleteIpvlanInterface(ep danmtypes.DanmEp) (error) { 
  return deleteEp(ep)
}

// DoesTargetContainerExist interrogates Docker whether the received CID belongs to an alive container, or it is outdated
func DoesTargetContainerExist(ep danmtypes.DanmEp) bool { 
  return doesTargetContainerExist(ep)
}

// FindByCid returns a map of Eps which belong to the same Pod
func FindByCid(client danmclientset.Interface, cid string)([]danmtypes.DanmEp, error) {
  result, err := client.DanmV1().DanmEps("").List(meta_v1.ListOptions{})
  if err != nil {
    log.Println("cannot get list of eps:" + err.Error())
    return nil, err
  }
  eplist := result.Items
  var ret = make([]danmtypes.DanmEp, 0)
  for _, ep := range eplist {
    if ep.Spec.CID == cid {
      ret = append(ret, ep)
    }
  }
  return ret, nil
}

// CidsByHost returns a map of Eps
// The Eps in the map are indexed with the name of the K8s host their Pods are running on
func CidsByHost(client danmclientset.Interface, host string)(map[string]danmtypes.DanmEp, error) {
  result, err := client.DanmV1().DanmEps("").List(meta_v1.ListOptions{})
  if err != nil {
    log.Println("cannot get list of eps")
    return nil, err
  }
  eplist := result.Items
  var ret = make(map[string]danmtypes.DanmEp, 0)
  for _, ep := range eplist {
    if ep.Spec.Host == host {
      ret[ep.Spec.CID] = ep
    }
  }
  return ret, nil
}

func AddIpvlanInterface(dnet *danmtypes.DanmNet, ep danmtypes.DanmEp) error {
  if ep.Spec.NetworkType != "ipvlan" {
    return nil
  }
  return createIpvlanInterface(dnet, ep)
}

func DetermineHostDeviceName(dnet *danmtypes.DanmNet) string {
  var device string
  isVlanDefined := (dnet.Spec.Options.Vlan!=0)
  isVxlanDefined := (dnet.Spec.Options.Vxlan!=0)
  if isVxlanDefined {
    device = "vx_" + dnet.Spec.NetworkID
  } else if isVlanDefined {
    vlanId := strconv.Itoa(dnet.Spec.Options.Vlan)
    device = dnet.Spec.NetworkID + "." + vlanId
  } else {
    device = dnet.Spec.Options.Device
  }
  return device
}