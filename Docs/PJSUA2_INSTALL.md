# pjsua2 Python Bindings — Installation Guide

pjsua2 is the C++/Python wrapper for PJSIP, the SIP stack used by siptty.

## System Dependencies (Ubuntu/Debian)

Install all required build packages in one shot:

```bash
sudo apt-get update
sudo apt-get install -y build-essential python3-dev swig \
    libasound2-dev libopus-dev libssl-dev
```

Or via the Makefile:

```bash
make install-deps
```

## Option 1 — pip install (may fail with SWIG ≥ 4.2)

Activate the project venv and install:

```bash
source .venv/bin/activate
pip install pjsua2
```

Or via the Makefile:

```bash
make install-pjsua2
```

### Known issue: SWIG 4.2.x incompatibility

The `pjsua2` PyPI package builds PJSIP from source using SWIG.  On Ubuntu
24.04 (which ships SWIG 4.2.0), the build fails with C++ template errors in
the generated `pjsua2_wrap.cpp`:

```
error: invalid use of incomplete type
  'struct swig::SwigPyMapIterator_T<...>'
```

Additionally, the PyPI package's `setup.py` expects to live inside the
pjproject source tree (`FileNotFoundError: '../../../../version.mak'`), so
`pip install pjsua2` currently fails on a clean system.

If you hit either of these errors, use Option 2 below.

## Option 2 — Build from pjproject source

This is the reliable path until a fixed wheel is available.

```bash
# 1. Clone pjproject
git clone --depth 1 --branch 2.14.1 https://github.com/pjsip/pjproject.git
cd pjproject

# 2. Configure & build the C libraries
./configure --enable-shared --with-opus=/usr --with-ssl=/usr
make dep && make -j$(nproc)

# 3. Build the Python bindings
cd pjsip-apps/src/swig/python
make          # runs SWIG then python3 setup.py build
make install  # installs into the active Python environment
```

**SWIG version requirement:** You need SWIG **4.1.x or earlier**.  SWIG 4.2.0
(the default on Ubuntu 24.04) generates broken map-iterator code for the
pjsua2 bindings.  Workarounds:

- Install SWIG 4.1.1 from source or a PPA before building.
- Use a pre-built `.whl` if one becomes available for your platform.
- Patch the generated `pjsua2_wrap.cpp` (fragile, not recommended).

## Option 3 — Pre-built wheel (when available)

If a `.whl` file is provided (e.g. built on CI):

```bash
source .venv/bin/activate
pip install pjsua2-*.whl
```

## Verify the installation

```bash
source .venv/bin/activate
python scripts/check_pjsua2.py
```

Expected output on success:

```
pjsua2 OK — PJSIP 2.14.1
```

Or in Python:

```python
import pjsua2 as pj
ep = pj.Endpoint()
ep.libCreate()
print(ep.libVersion().full)
ep.libDestroy()
```

You can also check the availability flag in application code:

```python
from siptty.engine import PJSUA2_AVAILABLE
print(f"pjsua2 available: {PJSUA2_AVAILABLE}")
```
