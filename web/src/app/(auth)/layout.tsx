'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  useEffect(() => {
    if (isAuthenticated) {
      router.replace('/discover');
    }
  }, [isAuthenticated, router]);

  return (
    <div className="flex min-h-screen">
      {/* Left: branding */}
      <div className="hidden w-1/2 items-center justify-center bg-gradient-to-br from-primary-500 to-primary-700 lg:flex">
        <div className="text-center text-white">
          <div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-3xl bg-white/20 backdrop-blur">
            <svg viewBox="0 0 24 24" className="h-12 w-12" fill="currentColor">
              <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
            </svg>
          </div>
          <h1 className="text-4xl font-bold">TrueConnect</h1>
          <p className="mt-3 text-lg text-white/80">Trust-based connections in Kazakhstan</p>
        </div>
      </div>

      {/* Right: form */}
      <div className="flex w-full items-center justify-center px-6 lg:w-1/2">
        <div className="w-full max-w-md">{children}</div>
      </div>
    </div>
  );
}
