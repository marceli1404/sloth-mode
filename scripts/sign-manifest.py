"""Owner-only helper. Run with an offline private key; never commit the PEM."""
import argparse, base64, hashlib, json
from pathlib import Path
from cryptography.hazmat.primitives.serialization import load_pem_private_key

def payload(m):
    return f"sloth-mode-update-v1\n{m['format']}\n{m['version']}\n{m['url']}\n{m['sha256']}\n{m['size']}\n".encode()

if __name__ == '__main__':
    p=argparse.ArgumentParser()
    p.add_argument('--signing-key',required=True,type=Path)
    p.add_argument('--version',required=True)
    p.add_argument('--installer',required=True,type=Path)
    p.add_argument('--url',required=True)
    p.add_argument('--output',required=True,type=Path)
    a=p.parse_args()
    if not a.url.startswith('https://'):
        p.error('Installer URL must be HTTPS')
    b=a.installer.read_bytes()
    m={'format':1,'version':a.version,'url':a.url,'sha256':hashlib.sha256(b).hexdigest(),'size':len(b)}
    k=load_pem_private_key(a.signing_key.read_bytes(),password=None)
    m['signature']=base64.b64encode(k.sign(payload(m))).decode()
    a.output.write_text(json.dumps(m,indent=2)+'\n')
    print(f'Signed manifest for {a.version} ({len(b)} bytes); NEVER publish the signing key.')
