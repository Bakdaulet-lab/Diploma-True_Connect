'use client';

import { useState } from 'react';
import { ShieldCheck, Upload, AlertCircle } from 'lucide-react';
import { useKycStatus, useSubmitKyc } from '@/hooks/api';
import { LoadingScreen } from '@/components/ui/common';

export default function KYCPage() {
  const { data: kycStatus, isLoading: isStatusLoading } = useKycStatus();
  const { mutate: submitKyc, isPending: isSubmitting } = useSubmitKyc();
  
  const [file, setFile] = useState<File | null>(null);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setFile(e.target.files[0]);
    }
  };

  const handleSubmit = () => {
    if (file) {
      submitKyc(file);
    }
  };

  if (isStatusLoading) {
    return <LoadingScreen />;
  }

  const isPending = kycStatus?.status === 'pending';
  const isApproved = kycStatus?.status === 'approved';

  return (
    <div className="container mx-auto max-w-2xl py-8 px-4">
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <ShieldCheck className="h-8 w-8 text-primary" />
            Identity Verification
          </h1>
          <p className="text-muted-foreground mt-2">
            Verify your identity to unlock all platform features and build trust.
          </p>
        </div>

        <div className="bg-card rounded-xl border shadow-sm p-6 flex flex-col items-center justify-center min-h-[300px] text-center">
          {!isPending && !isApproved ? (
            <div className="space-y-4 max-w-sm">
              <div className="mx-auto bg-primary/10 w-16 h-16 rounded-full flex items-center justify-center mb-6">
                <Upload className="h-8 w-8 text-primary" />
              </div>
              <h2 className="text-xl font-semibold">Upload ID Document</h2>
              <p className="text-sm text-muted-foreground">
                Please upload a clear photo of your passport, national ID card, or driver&apos;s license.
              </p>
              
              <div className="mt-6 border-2 border-dashed border-input rounded-xl p-8 hover:bg-muted/50 transition-colors cursor-pointer relative">
                <input 
                  type="file" 
                  className="absolute inset-0 w-full h-full opacity-0 cursor-pointer" 
                  accept="image/*"
                  onChange={handleFileUpload}
                  disabled={isSubmitting}
                />
                <span className="text-primary font-medium">
                  {file ? file.name : 'Click to browse or drag and drop'}
                </span>
              </div>
              
              <button 
                onClick={handleSubmit} 
                disabled={!file || isSubmitting}
                className="w-full mt-4 bg-primary text-primary-foreground py-2 px-4 rounded-md font-medium disabled:opacity-50"
              >
                {isSubmitting ? 'Uploading...' : 'Submit Document'}
              </button>
            </div>
          ) : (
            <div className="space-y-4 max-w-sm">
              <div className="mx-auto bg-green-100 w-16 h-16 rounded-full flex items-center justify-center mb-6">
                <ShieldCheck className="h-8 w-8 text-green-600" />
              </div>
              <h2 className="text-xl font-semibold">
                {isApproved ? 'Identity Verified' : 'Verification in Progress'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {isApproved 
                  ? 'Your identity has been successfully verified. You now have full access to all features.' 
                  : 'Your document has been submitted successfully. Our team will review it within 24-48 hours.'}
              </p>
              {!isApproved && (
                <div className="mt-8 p-4 bg-blue-50 text-blue-800 rounded-lg flex items-start gap-3 text-left">
                  <AlertCircle className="h-5 w-5 shrink-0 mt-0.5" />
                  <p className="text-sm">
                    You can continue using standard features. We&apos;ll notify you once verification is complete.
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
