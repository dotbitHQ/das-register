package timer

import (
	"encoding/json"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"testing"
)

func Test_ReverseSmtHasOldCreate(t *testing.T) {
	if err := txTimer.doReverseSmtTask(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalSign(t *testing.T) {
	signHandler := sign.LocalSign("a46f1213966ec1c4624557f4f84dee9f07f4faca684b184cc650ceddd21cecf7")
	bs, err := signHandler("0x3636373236663664323036343639363433613230392f4e4d4c6e5957486f78325a48714b6933627044714264467554384d703148317954717537626e4542733d")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bs))
}

func Test_GenWitness(t *testing.T) {

	builder := witness.NewReverseSmtBuilder()

	record := &witness.ReverseSmtRecord{
		Version:     1,
		Action:      "update",
		Signature:   "0x307836653663396233663732613437656231356565363333366537373665393161656538613264343339306638366666613232366230393934383262393535333761336266643936626236363363666530653136646161343330306331643865623362313361356438666563396339373131326462616661303762353265323966313162",
		SignType:    3,
		Address:     "0xc0d0087dA03480f9d7e7E1D76d2DCa4bb0A98B17",
		Proof:       "0x3078346334663030",
		PrevNonce:   0,
		PrevAccount: "",
		NextRoot:    []byte("df41487f90abe236cfc3b57cd269e50c75cd21a262ce2f8bd141fba8d28ef65d"),
		NextAccount: "test.bit",
	}
	bs, err := record.GenBytes()
	if err != nil {
		t.Fatal(err)
	}

	parseRecord, err := builder.FromBytes(bs)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*parseRecord)
}

func Test_ParseWitness(t *testing.T) {
	smtBuilder := witness.NewReverseSmtBuilder()

	tx := &types.Transaction{}

	txJson := `
{
    "version":"0x0",
    "cell_deps":[
        {
            "out_point":{
                "tx_hash":"0x6678cf2da360945b031170b2e23e776b684eac29b3f9c1f8a2cf62cb88f2a4f7",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0xdf4337204e1fb77c9deac20afe0e4fcb83a8d986db0f7250874485dffd48e66e",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0xd819d634d5593d1c7f22d8f954fd743c50eaea427fe669595b7f7a1109bbac6f",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0x77cdb8d076e3780ef46c42e8f473e9ec2ea1d9521e1cf8ee0db9efb01671d341",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0xf249a946f1302c34d63d437eaf345ce77b96c91f142cef3c356ec16f0ecc3f34",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0x56ae7c87f1fdb7255b5cbc918187aea22a2c73e6ee651e61ae0a327703c8d63f",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0x8ffa409ba07d74f08f63c03f82b7428d36285fe75b2173fc2476c0f7b80c707a",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0x9e0823959e5b76bd010cc503964cced4f8ae84f3b03e94811b083f9765534ff1",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0xa706f46e58e355a6d29d7313f548add21b875639ea70605d18f682c1a08740d6",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0x747411fb3914dd7ca5488a0762c6f4e76f56387e83bcbb24e3a01afef1d5a5b4",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0x209b35208da7d20d882f0871f3979c68c53981bcc4caa71274c035449074d082",
                "index":"0x0"
            },
            "dep_type":"code"
        },
        {
            "out_point":{
                "tx_hash":"0xf8de3bb47d055cdf460d93a2a6e1b05f7432f9777c8c474abf4eec1d4aee5d37",
                "index":"0x0"
            },
            "dep_type":"dep_group"
        },
        {
            "out_point":{
                "tx_hash":"0xa7ff448225fc131d657af882a3f97a8219be230d7e25d070a9282de89302c640",
                "index":"0x0"
            },
            "dep_type":"code"
        }
    ],
    "header_deps":[
        "0x6293045bb57025c28e724217420837d7a7d2a4f71f1364f58e6681353b2d0ddd"
    ],
    "inputs":[
        {
            "since":"0x0",
            "previous_output":{
                "tx_hash":"0x17789c41d0bd890fe305a3dc50ce92911111518ec3b29145bf6ecb868052ed36",
                "index":"0x0"
            }
        },
        {
            "since":"0x0",
            "previous_output":{
                "tx_hash":"0x5b5406b6bd97a0dd8a3acf123e9cc9844c99ea4b4bd83655e1d171dbc88e3efb",
                "index":"0x0"
            }
        }
    ],
    "outputs":[
        {
            "capacity":"0x4a817c800",
            "lock":{
                "code_hash":"0xf1ef61b6977508d9ec56fe43399a01e576086a76cf0f7c687d1418335e8c401f",
                "hash_type":"type",
                "args":"0x"
            },
            "type":{
                "code_hash":"0x8041560ab6bd812c4523c824f2dcf5843804a099cb2f69fcbd57c8afcef2ed5f",
                "hash_type":"type",
                "args":"0x"
            }
        },
        {
            "capacity":"0xe8d4a4e8f0",
            "lock":{
                "code_hash":"0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8",
                "hash_type":"type",
                "args":"0xda44ed9db97056a06e471d3a1b6a1b82219e7232"
            },
            "type":null
        }
    ],
    "outputs_data":[
        "0x307864663431343837663930616265323336636663336235376364323639653530633735636432316132363263653266386264313431666261386432386566363564",
        "0x"
    ],
    "witnesses":[
        "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "0x55000000100000005500000055000000410000000ca1784f83c2f9c8ac19e112b20516558f0bfee1ae4084722ad4480ebe7355d76891e5ee980bcff9f04f965e989a9b55010e16736b2750a9cd158cb93a60f16400",
        "0x646173000000002f0000000c0000002a0000001a0000007570646174655f726576657273655f7265636f72645f726f6f740100000000",
        "0x6461730a0000000400000001000000060000007570646174658400000030783865343765626665333664373135336233306239343831383931653930373632613139626233376466323262336231633530313038633334343166623132656530353362353737346533333466396331393832353665306432396437623136323036633432613331643532636363303764343462376432393935323233626536316201000000032a000000307863306430303837644130333438306639643765374531443736643244436134626230413938423137080000003078346334663030000000000000000020000000df41487f90abe236cfc3b57cd269e50c75cd21a262ce2f8bd141fba8d28ef65d08000000746573742e626974",
        "0x64617368000000dd0200001000000011000000e901000001d801000038000000580000007800000098000000b8000000d8000000f80000001801000038010000580100007801000098010000b80100001106d9eaccde0995a7e07e80dd0ce7509f21752538dfdd1ee2526d24574846b10fbff871dd05aee1fda2be38786ad21d52a2765c6025d1ef6927d761d51a3cd14ff58f2c76b4ac26fdf675aa82541e02e4cf896279c6d6982d17b959788b2f0c08d1cdc6ab92d9cabe0096a2c7642f73d0ef1b24c94c43f21c6c3a32ffe0bb5e6c8441233f00741955f65e476721a1a5417997c1e4368801c99c7f617f8b754467d48c0911e406518de2116bd91c6af37c05f1db23334ca829d2af3042427e449438124abdf4cbbfd61065e8b64523172bef5eefe27cb769c40acaf036aa89c200000000000000000000000000000000000000000000000000000000000000001a3f02aa89651a18112f0c21d0ae370a86e13f6a060c378184cd859a7bb6520361711416468fa5211ead5f24c6f3efadfbbc332274c5d40e50c6feadcb5f96068bb0413701cdd2e3a661cc8914e6790e16d619ce674930671e695807274bd14c4fd085557b4ef857b0577723bbf0a2e94081bbe3114de847cd9db01abaeb4f4e8041560ab6bd812c4523c824f2dcf5843804a099cb2f69fcbd57c8afcef2ed5ff40000001c000000400000006400000088000000ac000000d0000000209b35208da7d20d882f0871f3979c68c53981bcc4caa71274c035449074d08200000000747411fb3914dd7ca5488a0762c6f4e76f56387e83bcbb24e3a01afef1d5a5b4000000000000000000000000000000000000000000000000000000000000000000000000000000008ffa409ba07d74f08f63c03f82b7428d36285fe75b2173fc2476c0f7b80c707a000000009e0823959e5b76bd010cc503964cced4f8ae84f3b03e94811b083f9765534ff100000000a706f46e58e355a6d29d7313f548add21b875639ea70605d18f682c1a08740d600000000",
        "0x646173700000002800000010000000180000002000000000c817a80400000000e1f505000000001027000000000000",
        "0x646173740000002400000061336d521b8c43e3b38686c3923f05051a1e0416ff556907b37a6ee06ce84246"
    ]
}
`

	var inTransaction = &struct {
		Version     hexutil.Uint    `json:"version"`
		HeaderDeps  []types.Hash    `json:"header_deps"`
		OutputsData []hexutil.Bytes `json:"outputs_data"`
		Witnesses   []hexutil.Bytes `json:"witnesses"`
	}{}

	if err := json.Unmarshal([]byte(txJson), inTransaction); err != nil {
		t.Fatal(err)
	}
	_ = gconv.Struct(inTransaction, tx)

	txReverseSmtRecord, err := smtBuilder.FromTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(txReverseSmtRecord)
}
