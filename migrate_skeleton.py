
import re

path = 'internal/apps/skeleton/template/server/server_integration_test.go'
with open(path, 'r', encoding='utf-8') as f:
    content = f.read()

idx = content.find('func TestSkeletonTemplateServer_ShutdownIdempotent')
if idx >= 0:
    print('Found at index', idx)
    end = content.find('\nfunc ', idx + 1)
    print(repr(content[idx:end]))
else:
    print('NOT FOUND')
