'use client';

import Link from 'next/link';
import { Edit, Shield, Settings, Camera } from 'lucide-react';
import { useMyProfile, useMyPhotos, useCurrentUser } from '@/hooks/api';
import { LoadingScreen, Avatar, TrustBadge } from '@/components/ui/common';
import { calculateAge } from '@/lib/utils';

export default function ProfilePage() {
  const { data: profile, isLoading: profileLoading } = useMyProfile();
  const { data: photos } = useMyPhotos();
  const { data: user } = useCurrentUser();

  if (profileLoading) return <LoadingScreen />;

  if (!profile) {
    return (
      <div className="flex flex-col items-center py-16 text-center">
        <Camera className="h-12 w-12 text-gray-400 mb-4" />
        <h2 className="text-xl font-bold text-gray-900">Set up your profile</h2>
        <p className="mt-2 text-sm text-gray-500">Complete your profile to start connecting</p>
        <Link href="/profile/edit" className="btn-primary mt-4">
          Create Profile
        </Link>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">My Profile</h1>
        <div className="flex gap-2">
          <Link href="/profile/edit" className="btn-secondary gap-2">
            <Edit className="h-4 w-4" />
            Edit
          </Link>
          <Link href="/settings" className="btn-ghost">
            <Settings className="h-5 w-5" />
          </Link>
        </div>
      </div>

      <div className="card">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Avatar src={profile.avatar_url} name={profile.display_name} size="xl" />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h2 className="text-xl font-bold text-gray-900">{profile.display_name}</h2>
              {profile.birth_date && (
                <span className="text-lg text-gray-500">{calculateAge(profile.birth_date)}</span>
              )}
            </div>
            {profile.city && <p className="text-sm text-gray-500">{profile.city}</p>}
            <div className="mt-2 flex items-center gap-3">
              {user && <TrustBadge score={user.trust_score} />}
              {user && (
                <span className="text-xs text-gray-400">
                  {user.verification_level === 'verified' ? '✓ Verified' : 'Not verified'}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Bio */}
        {profile.bio && (
          <div className="mt-6">
            <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">About</h3>
            <p className="mt-2 text-sm text-gray-700 whitespace-pre-wrap">{profile.bio}</p>
          </div>
        )}

        {/* Details */}
        <div className="mt-6 grid grid-cols-2 gap-4">
          {profile.gender && (
            <div>
              <p className="text-xs text-gray-400 uppercase">Gender</p>
              <p className="text-sm font-medium text-gray-700 capitalize">{profile.gender}</p>
            </div>
          )}
          {profile.looking_for && (
            <div>
              <p className="text-xs text-gray-400 uppercase">Looking for</p>
              <p className="text-sm font-medium text-gray-700 capitalize">{profile.looking_for}</p>
            </div>
          )}
        </div>

        {/* Photos */}
        {photos && photos.length > 0 && (
          <div className="mt-6">
            <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-3">Photos</h3>
            <div className="grid grid-cols-3 gap-2">
              {photos.map((photo) => (
                <img
                  key={photo.id}
                  src={photo.url}
                  alt=""
                  className="aspect-square w-full rounded-xl object-cover"
                />
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="mt-4 flex gap-3">
        <Link href="/kyc" className="btn-secondary flex-1 gap-2">
          <Shield className="h-4 w-4" />
          KYC Verification
        </Link>
      </div>
    </div>
  );
}
