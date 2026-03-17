'use client';

import { useParams } from 'next/navigation';
import { useState } from 'react';
import { MapPin, Flag } from 'lucide-react';
import { useProfile, useReputation, useReportUser } from '@/hooks/api';
import { LoadingScreen, ErrorMessage, Avatar, TrustBadge } from '@/components/ui/common';
import { calculateAge } from '@/lib/utils';
import toast from 'react-hot-toast';

export default function UserProfilePage() {
  const params = useParams<{ id: string }>();
  const { data: profile, isLoading, error } = useProfile(params.id);
  const { data: reputation } = useReputation(params.id);
  const reportMutation = useReportUser();
  const [showReport, setShowReport] = useState(false);
  const [reportReason, setReportReason] = useState('');

  const handleReport = () => {
    if (!reportReason.trim()) return;
    reportMutation.mutate(
      { reported_id: params.id, reason: reportReason.trim() },
      {
        onSuccess: () => {
          toast.success('Report submitted');
          setShowReport(false);
          setReportReason('');
        },
        onError: () => toast.error('Failed to submit report'),
      }
    );
  };

  if (isLoading) return <LoadingScreen />;
  if (error || !profile) return <ErrorMessage message="Profile not found" />;

  return (
    <div>
      <div className="card">
        <div className="flex items-center gap-4">
          <Avatar src={profile.avatar_url} name={profile.display_name} size="xl" />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h2 className="text-xl font-bold text-gray-900">{profile.display_name}</h2>
              {profile.birth_date && (
                <span className="text-lg text-gray-500">{calculateAge(profile.birth_date)}</span>
              )}
            </div>
            {profile.city && (
              <p className="flex items-center gap-1 text-sm text-gray-500">
                <MapPin className="h-4 w-4" />
                {profile.city}
              </p>
            )}
            {reputation && (
              <div className="mt-2 flex items-center gap-2">
                <TrustBadge score={reputation.score} />
                <span className="text-xs text-gray-400">
                  {reputation.rating_count} ratings
                </span>
              </div>
            )}
          </div>
        </div>

        {profile.bio && (
          <div className="mt-6">
            <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">About</h3>
            <p className="mt-2 text-sm text-gray-700 whitespace-pre-wrap">{profile.bio}</p>
          </div>
        )}

        {profile.photos && profile.photos.length > 0 && (
          <div className="mt-6">
            <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-3">Photos</h3>
            <div className="grid grid-cols-3 gap-2">
              {profile.photos.map((photo) => (
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

        <div className="mt-6 border-t border-gray-100 pt-4">
          <button
            onClick={() => setShowReport(true)}
            className="flex items-center gap-2 text-sm text-gray-400 hover:text-red-500 transition-colors"
          >
            <Flag className="h-4 w-4" />
            Report user
          </button>
        </div>
      </div>

      {/* Report Dialog */}
      {showReport && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-6">
            <h3 className="text-lg font-semibold text-gray-900">Report User</h3>
            <textarea
              value={reportReason}
              onChange={(e) => setReportReason(e.target.value)}
              placeholder="Describe the issue..."
              rows={3}
              className="input-field mt-3"
            />
            <div className="mt-4 flex gap-3">
              <button onClick={() => setShowReport(false)} className="btn-secondary flex-1">
                Cancel
              </button>
              <button
                onClick={handleReport}
                disabled={!reportReason.trim() || reportMutation.isPending}
                className="btn-primary flex-1 bg-red-500 hover:bg-red-600"
              >
                Report
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
