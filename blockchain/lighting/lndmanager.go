package lighting

import (
	"encoding/hex"
	"io/ioutil"
	"os"

	"github.com/astaxie/beego/logs"
	"github.com/jualy007/GoTF/config"
	"github.com/lightningnetwork/lnd/lnrpc"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Adapter struct {
	cc   *grpc.ClientConn
	lc   lnrpc.LightningClient
	meta metadata.MD
}

func NewAdapter(info *config.LndInfo) (adapter Adapter, err error) {
	creds, _ := credentials.NewClientTLSFromFile(info.Cert, "")

	adapter.cc, err = grpc.Dial(info.Address, grpc.WithTransportCredentials(creds))

	if err != nil {
		logs.Error("gRPC Connection %v Failed!!!", info.Address)
		return adapter, err
	}

	logs.Info("Start init Lightning Network Client......")
	adapter.lc = lnrpc.NewLightningClient(adapter.cc)

	fh, openerr := os.Open(info.Macaroon)
	if openerr != nil {
		logs.Error("Open Macaroon File %v Failed!!!", info.Macaroon)
	}

	defer fh.Close()
	fd, _ := ioutil.ReadAll(fh)

	macaroon := hex.EncodeToString(fd)
	adapter.meta = metadata.Pairs("macaroon", macaroon)

	return adapter, nil
}

func (adapter *Adapter) Close() {
	if adapter.cc != nil {
		defer adapter.cc.Close()
	}
}

func (adapter *Adapter) QueryRoute(payreq string) *lnrpc.QueryRoutesResponse {
	ctx := metadata.NewOutgoingContext(context.Background(), adapter.meta)

	payreqStr := adapter.DecodePayReq(payreq)

	routeinfo, err := adapter.lc.QueryRoutes(ctx, &lnrpc.QueryRoutesRequest{
		Amt:        payreqStr.NumSatoshis,
		PubKey:     payreqStr.Destination,
		RouteHints: payreqStr.RouteHints,
	})

	if err != nil {
		logs.Error("Query Route Failed For Pubkey %v and Amt %v!!!", payreqStr.Destination, payreqStr.NumSatoshis)
		return nil
	}

	return routeinfo
}

func (adapter *Adapter) GetInfo() (info *lnrpc.GetInfoResponse, err error) {
	ctx := metadata.NewOutgoingContext(context.Background(), adapter.meta)
	return adapter.lc.GetInfo(ctx, &lnrpc.GetInfoRequest{})
}

func (adapter Adapter) DecodePayReq(payreq string) *lnrpc.PayReq {
	ctx := metadata.NewOutgoingContext(context.Background(), adapter.meta)

	payreqStr, err := adapter.lc.DecodePayReq(ctx, &lnrpc.PayReqString{
		PayReq: payreq,
	})

	if err != nil {
		logs.Error("Decode Lightning Network Payment %v Failed!!!", payreq)
		return nil
	}

	return payreqStr
}
