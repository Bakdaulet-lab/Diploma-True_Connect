'use client';

import Link from 'next/link';
import { Edit, Shield, Settings, Camera, Trash2, Plus } from 'lucide-react';
import { useMyProfile, useMyPhotos, useCurrentUser, useUploadPhoto, useDeletePhoto } from '@/hooks/api';
import { LoadingScreen, Avatar, TrustBadge } from '@/components/ui/common';
import { calculateAge } from '@/lib/utils';

export default function ProfilePage() {
  const { data: profile, isLoading: profileLoading } = useMyProfile();
  const { data: photos } = useMyPhotos();
  const { data: user } = useCurrentUser();
  const { mutate: uploadPhoto, isPending: isUploading } = useUploadPhoto();
  const { mutate: deletePhoto } = useDeletePhoto();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      uploadPhoto(file);
    }
  };

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
              {profile.trust_score !== undefined && (
                <div className="flex items-center gap-2">
                  <TrustBadge score={profile.trust_score} />
                  {profile.badge && (
                    <span className="text-xs font-semibold px-2 py-1 rounded-full bg-purple-100 text-purple-700">
                      {profile.badge}
                    </span>
                  )}
                </div>
              )}
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
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">Photos</h3>
            <label className="btn-ghost cursor-pointer text-xs p-2 shrink-0 h-auto">
              {isUploading ? 'Uploading...' : <><Plus className="h-4 w-4 mr-1"/> Add Photo</>}
              <input type="file" accept="image/*" className="hidden" disabled={isUploading} onChange={handleFileChange} />
            </label>
          </div>
          
          {(!photos || photos.length === 0) ? (
            <div className="bg-gray-50 rounded-xl p-8 text-center border-2 border-dashed border-gray-200">
              <p className="text-sm text-gray-500 mb-2">No photos uploaded yet</p>
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-2">
              {photos.map((photo) => (
                <div key={photo.id} className="relative group aspect-square">
                  <img
                    src={photo.url}
                    alt=""
                    className="w-full h-full rounded-xl object-cover"
                  />
                  <button
                    onClick={() => deletePhoto(photo.id)}
                    className="absolute top-1 right-1 bg-red-500 text-white p-1.5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
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
