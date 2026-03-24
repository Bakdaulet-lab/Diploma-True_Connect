'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { Sidebar, BottomNav } from '@/components/layout/sidebar';
import { Spinner } from '@/components/ui/common';

export default function MainLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace('/login');
    }
  }, [isAuthenticated, router]);

  if (!isAuthenticated) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Sidebar />
      <BottomNav />
      <main className="pb-20 md:ml-64 md:pb-0">
        <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6">{children}</div>
      </main>
    </div>
  );
}
