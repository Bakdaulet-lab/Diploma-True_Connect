'use client';

import { useState, useEffect } from 'react';
import { useMyProfile, useSettings, useUpdateSettings } from '@/hooks/api';
import { LoadingScreen } from '@/components/ui/common';
import type { UserSettings } from '@/types';

export default function SettingsPage() {
  const { data: profile, isLoading: isProfileLoading } = useMyProfile();
  const { data: settings, isLoading: isSettingsLoading } = useSettings();
  const { mutate: updateSettings, isPending: isUpdating } = useUpdateSettings();
  
  const [localSettings, setLocalSettings] = useState<UserSettings | undefined>(settings);

  useEffect(() => {
    if (settings) {
      setLocalSettings(settings);
    }
  }, [settings]);

  if (isProfileLoading || isSettingsLoading) {
    return <LoadingScreen />;
  }

  const handleToggle = (key: keyof UserSettings) => {
    if (!localSettings) return;
    const newValue = !localSettings[key];
    setLocalSettings({ ...localSettings, [key]: newValue } as UserSettings);
    updateSettings({ [key]: newValue });
  };

  const handleRangeChange = (key: keyof UserSettings, value: number) => {
    if (!localSettings) return;
    setLocalSettings({ ...localSettings, [key]: value } as UserSettings);
    // In a real app we might debounce this to avoid too many requests
    updateSettings({ [key]: value });
  };

  return (
    <div className="container mx-auto max-w-2xl py-8 px-4">
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
          <p className="text-muted-foreground mt-2">
            Manage your account settings and preferences.
          </p>
        </div>

        <div className="bg-card rounded-xl border shadow-sm p-6">
          <h2 className="text-xl font-semibold mb-4">Profile Information</h2>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Name</label>
              <div className="p-2 bg-muted rounded-md">{profile?.display_name || 'Not set'}</div>
            </div>
            <div>
              <label className="text-sm font-medium mb-1 block">Bio</label>
              <div className="p-2 bg-muted rounded-md min-h-[60px]">{profile?.bio || 'No bio'}</div>
            </div>
          </div>
        </div>

        <div className="bg-card rounded-xl border shadow-sm p-6">
          <h2 className="text-xl font-semibold mb-4">Account Preferences</h2>
          <p className="text-sm text-muted-foreground mb-4">
            Update your appearance and notification settings here.
          </p>
          
          {localSettings && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-medium">Show Online Status</div>
                  <div className="text-sm text-muted-foreground">Let others see when you are active</div>
                </div>
                <button 
                  onClick={() => handleToggle('show_online_status')}
                  className={`h-6 w-11 rounded-full relative transition-colors ${localSettings.show_online_status ? 'bg-primary' : 'bg-secondary'}`}
                  disabled={isUpdating}
                >
                  <span className={`block h-4 w-4 rounded-full bg-white absolute top-1 transition-transform ${localSettings.show_online_status ? 'left-6' : 'left-1'}`} />
                </button>
              </div>
              
              <div className="flex items-center justify-between">
                <div>
                  <div className="font-medium">Push Notifications</div>
                  <div className="text-sm text-muted-foreground">Receive updates on matches and messages</div>
                </div>
                <button 
                  onClick={() => handleToggle('push_notifications')}
                  className={`h-6 w-11 rounded-full relative transition-colors ${localSettings.push_notifications ? 'bg-primary' : 'bg-secondary'}`}
                  disabled={isUpdating}
                >
                  <span className={`block h-4 w-4 rounded-full bg-white absolute top-1 transition-transform ${localSettings.push_notifications ? 'left-6' : 'left-1'}`} />
                </button>
              </div>

              <div>
                <div className="flex justify-between mb-2">
                  <div className="font-medium">Maximum Distance</div>
                  <div className="text-sm">{localSettings.max_distance_km} {localSettings.distance_unit || 'km'}</div>
                </div>
                <input 
                  type="range" 
                  className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer" 
                  min="1" max="100" 
                  value={localSettings.max_distance_km || 10} 
                  onChange={(e) => handleRangeChange('max_distance_km', parseInt(e.target.value))}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
