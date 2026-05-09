'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { useQueryClient } from '@tanstack/react-query'; // Добавлен импорт
import { useMyProfile, useUpdateProfile } from '@/hooks/api';
import { LoadingScreen } from '@/components/ui/common';
import type { ProfileUpsert } from '@/types';
import toast from 'react-hot-toast';

export default function EditProfilePage() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data: profile, isLoading } = useMyProfile();
  const { mutate: updateProfile, isPending } = useUpdateProfile();

  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<ProfileUpsert>();

  const currentLat = watch('latitude');
  const currentLon = watch('longitude');

  const handleUpdateLocation = () => {
    if (!navigator.geolocation) {
      toast.error('Geolocation is not supported by your browser');
      return;
    }

    toast.loading('Fetching location...', { id: 'loc' });
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setValue('latitude', position.coords.latitude);
        setValue('longitude', position.coords.longitude);
        toast.success('Location updated!', { id: 'loc' });
      },
      (error) => {
        toast.error('Could not get precise location', { id: 'loc' });
        console.error(error);
      }
    );
  };

  useEffect(() => {
    if (profile) {
      reset({
        display_name: profile.display_name || '',
        bio: profile.bio || '',
        gender: profile.gender || '',
        birth_date: profile.birth_date ? profile.birth_date.split('T')[0] : '',
        city: profile.city || '',
        latitude: profile.latitude,
        longitude: profile.longitude,
      });
    }
  }, [profile, reset]);

  if (isLoading) return <LoadingScreen />;

  const onSubmit = (data: ProfileUpsert) => {
    // Очищаем пустые строки, чтобы бэкенд не ругался на валидацию
    const cleanData = {
      ...data,
      gender: data.gender === '' ? undefined : data.gender,
      looking_for: (data.looking_for === '' || data.looking_for === 'both') ? undefined : data.looking_for,
      birth_date: data.birth_date === '' ? undefined : data.birth_date,
    };

    updateProfile(cleanData, {
      onSuccess: async () => {
        // ЖДЕМ, пока кэш полностью сбросится, и только потом делаем редирект
        await queryClient.invalidateQueries({ queryKey: ['myProfile'] });
        
        toast.success('Profile updated successfully');
        router.push('/profile');
      },
      onError: () => {
        toast.error('Failed to update profile');
      }
    });
  };

  return (
    <div className="max-w-2xl mx-auto py-8">
      <h1 className="text-2xl font-bold mb-6 text-gray-900">Edit Profile</h1>
      
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 bg-white p-6 rounded-2xl shadow-sm border border-gray-100">
        <div>
          <label className="block text-sm font-medium mb-2 text-gray-700">Display Name</label>
          <input
            {...register('display_name', { required: 'Display name is required', minLength: { value: 2, message: 'Minimum 2 characters' } })}
            className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-primary-500 focus:ring-primary-500"
            placeholder="Your name"
          />
          {errors.display_name && <p className="text-red-500 text-sm mt-1">{errors.display_name.message}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2 text-gray-700">Birth Date</label>
          <input
            type="date"
            {...register('birth_date')}
            className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-primary-500 focus:ring-primary-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-2 text-gray-700">Bio</label>
          <textarea
            {...register('bio')}
            rows={4}
            className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-primary-500 focus:ring-primary-500"
            placeholder="Tell us about yourself..."
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2 text-gray-700">Gender</label>
            <select
              {...register('gender')}
              className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-primary-500 focus:ring-primary-500 bg-white"
            >
              <option value="">Select...</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
              <option value="other">Other</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-2 text-gray-700">Looking For</label>
            <select
              {...register('looking_for')}
              className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-primary-500 focus:ring-primary-500 bg-white"
            >
              <option value="">Select...</option>
              <option value="male">Men</option>
              <option value="female">Women</option>
              <option value="both">Everyone</option>
            </select>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2 text-gray-700">City</label>
            <input
              {...register('city')}
              className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-primary-500 focus:ring-primary-500 mb-2"
              placeholder="Where do you live?"
            />
            {currentLat && currentLon && (
              <p className="text-xs text-green-600 mb-2 font-medium">
                ✓ Precise coordinates captured (GPS)
              </p>
            )}
            <button
              type="button"
              onClick={handleUpdateLocation}
              className="text-sm px-4 py-2 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded-lg font-medium transition-colors"
            >
              Update Precise Location
            </button>
          </div>
        </div>

        <div className="pt-4 flex justify-end gap-3 flex-wrap">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-6 py-3 rounded-xl border border-gray-300 font-medium text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="px-6 py-3 rounded-xl bg-primary-500 text-white font-medium hover:bg-primary-600 transition-colors disabled:opacity-50"
          >
            {isPending ? 'Saving...' : 'Save Profile'}
          </button>
        </div>
      </form>
    </div>
  );
}