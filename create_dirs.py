import os
base = r'c:\Users\Пользователь\Desktop\Mobile-dev\iter-2\web'
dirs = [
    'public',
    'src/app/(auth)/login',
    'src/app/(auth)/register', 
    'src/app/(main)/discover',
    'src/app/(main)/matches',
    'src/app/(main)/chat/[id]',
    'src/app/(main)/feed',
    'src/app/(main)/profile/[id]',
    'src/app/(main)/settings',
    'src/app/(main)/kyc',
    'src/components/ui',
    'src/components/auth',
    'src/components/discovery',
    'src/components/chat',
    'src/components/feed',
    'src/components/profile',
    'src/components/layout',
    'src/hooks',
    'src/lib',
    'src/store',
    'src/types',
    'src/styles',
]
for d in dirs:
    path = os.path.join(base, d)
    os.makedirs(path, exist_ok=True)
    print(f'Created: {path}')
print('Done!')
