import os
import traceback
import requests

from xml.etree import ElementTree
from ndn.app import NDNApp
from ndn.encoding.name.Component import to_str, to_number, get_type, from_number, from_segment, from_version, TYPE_SEGMENT, TYPE_VERSION
from ndn.types import InterestNack

app = NDNApp()

SERVE_PREFIX = '/demo/fileserve'
DATA_VERSION = 1
BASE_DIR = 'public/'
SEGMENT_SIZE = 4400

serve_segs = len(SERVE_PREFIX.split('/'))-1
vers_seg = from_number(DATA_VERSION, typ=TYPE_VERSION)

basequery = 'http://export.arxiv.org/api/query?search_query=doi:'

@app.route(SERVE_PREFIX)
def on_interest(name, interest_param, application_param):
    try:
        print("got: ", [to_str(n) for n in name])
        if to_str(name[-1]) == "metadata":
            # serving_prefix / name / metadata
            filename = BASE_DIR + '/'.join([to_str(n) for n in name[serve_segs:-1]])
            print("discovery request, checking", filename)
            if not os.path.isfile(filename):
                if to_str(name[serve_segs]) == "doi":
                    doi = '/'.join([to_str(n) for n in name[serve_segs+1:-1]])
                    queryxml = requests.get(basequery + doi).text
                    pdfurl = ElementTree.fromstring(queryxml).find(".//*[@title='pdf']").get('href')
                    pdf = requests.get(pdfurl).content

                    os.makedirs(os.path.dirname(filename), 0o777, True)
                    with open(filename, 'wb')  as f:
                        f.write(pdf)
                else:
                    raise InterestNack(150)
            first = name
            first.append(vers_seg)
            first.append(from_segment(0))
            app.put_data(first, freshness_period=100)
            return
        elif get_type(name[-1]) == TYPE_SEGMENT and get_type(name[-2]) == TYPE_VERSION:
            # serving_prefix / name / version / chunk
            if to_number(name[-2]) != DATA_VERSION:
                raise InterestNack(150)
            filename = BASE_DIR + '/'.join([to_str(n) for n in name[serve_segs:-2]])
            chunk = to_number(name[-1])
        else:
            raise InterestNack(150)

        with open(filename, 'rb') as f:
            f.seek(chunk * SEGMENT_SIZE, 0)
            b = f.read(SEGMENT_SIZE)
            pos = f.tell()
            f.seek(0, os.SEEK_END)
            end = f.tell()

            final_block = f.tell() // SEGMENT_SIZE
            final_block_id = from_segment(final_block)
            if final_block < chunk:
                raise InterestNack(150)

            app.put_data(
                name,
                content=b,
                freshness_period=100000,
                final_block_id=final_block_id)
    except Exception as e:
        print("Exception:", str(e))
        app.put_data(
            name,
            content=app.get_original_packet_value(name),
            content_type=3)

app.run_forever()
