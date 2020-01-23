import os
import sys

from ndn.app import NDNApp
from ndn.encoding.name.Component import from_segment, to_number, to_str
from ndn.encoding.ndn_format_0_3 import ContentType

app = NDNApp()

async def main():
    content = bytes()
    try:
        name = sys.argv[1] + "/metadata"
        print("sending", name, file=sys.stderr)
        dname, meta, mcontent = await app.express_interest(
            name,
            must_be_fresh=True,
            can_be_prefix=True,
            lifetime=10000)

        print("got", [to_str(n) for n in dname], file=sys.stderr)
        name = dname[:-3] + dname[-2:]

        while True:
            print("sending", [to_str(n) for n in name], file=sys.stderr)
            dname, meta, scontent = await app.express_interest(
                name,
                lifetime=10000)
            if meta.content_type == ContentType.NACK:
                print("got NACK")
                break

            print("got", [to_str(n) for n in dname], file=sys.stderr)
            content = content + scontent
            seg = to_number(dname[-1])

            if to_number(meta.final_block_id) == seg:
                break
            else:
                name = name[:-1]
                name.append(from_segment(seg+1))

    except Exception as e:
        print("Exception:", str(e), file=sys.stderr)
    finally:
        if len(content) > 0:
            if len(sys.argv) > 2:
                with open(sys.argv[2], 'wb') as f:
                    f.write(content)
            else:
                print(content)
        app.shutdown()

app.run_forever(after_start=main())
