import { zodResolver } from '@hookform/resolvers/zod';
import { Edit, Loader2 } from 'lucide-react';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { ISO, UpdateISORequest } from '../types/iso';

// Schema for updating ISOs - all fields are optional
const updateISOSchema = z.object({
  name: z.string().min(1, 'Name is required').optional(),
  version: z.string().min(1, 'Version is required').optional(),
  arch: z.enum(['x86_64', 'aarch64', 'arm64', 'i686']).optional(),
  edition: z.string().optional(),
  download_url: z
    .string()
    .url('Must be a valid URL')
    .optional()
    .or(z.literal('')),
  checksum_url: z
    .string()
    .url('Must be a valid URL')
    .optional()
    .or(z.literal('')),
  checksum_type: z.enum(['sha256', 'sha512', 'md5']).optional(),
});

type UpdateISOFormData = z.infer<typeof updateISOSchema>;

interface EditIsoModalProps {
  iso: ISO;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (id: string, request: UpdateISORequest) => Promise<void>;
}

export function EditIsoModal({
  iso,
  open,
  onOpenChange,
  onSubmit,
}: EditIsoModalProps) {
  const canEditURLs = iso.status === 'failed';

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
    watch,
    setValue,
  } = useForm<UpdateISOFormData>({
    resolver: zodResolver(updateISOSchema),
  });

  // Reset form when ISO changes or dialog opens
  useEffect(() => {
    if (open) {
      reset({
        name: iso.name,
        version: iso.version,
        arch: iso.arch as 'x86_64' | 'aarch64' | 'arm64' | 'i686',
        edition: iso.edition || '',
        download_url: iso.download_url,
        checksum_url: iso.checksum_url || '',
        checksum_type: (iso.checksum_type || 'sha256') as
          | 'sha256'
          | 'sha512'
          | 'md5',
      });
    }
  }, [iso, open, reset]);

  const archValue = watch('arch');
  const checksumTypeValue = watch('checksum_type');

  const onFormSubmit = async (data: UpdateISOFormData) => {
    try {
      // Only send changed fields
      const updateRequest: UpdateISORequest = {};
      if (data.name && data.name !== iso.name) updateRequest.name = data.name;
      if (data.version && data.version !== iso.version)
        updateRequest.version = data.version;
      if (data.arch && data.arch !== iso.arch) updateRequest.arch = data.arch;
      if (data.edition !== undefined && data.edition !== iso.edition)
        updateRequest.edition = data.edition;

      // Only include URL fields for failed ISOs
      if (canEditURLs) {
        if (data.download_url && data.download_url !== iso.download_url)
          updateRequest.download_url = data.download_url;
        if (
          data.checksum_url !== undefined &&
          data.checksum_url !== iso.checksum_url
        )
          updateRequest.checksum_url = data.checksum_url || undefined;
        if (
          data.checksum_type &&
          data.checksum_type !== iso.checksum_type
        )
          updateRequest.checksum_type = data.checksum_type;
      }

      await onSubmit(iso.id, updateRequest);
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to update ISO:', error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            <div className="flex items-center gap-2">
              <Edit className="w-5 h-5" />
              Edit ISO
            </div>
          </DialogTitle>
          <DialogDescription>
            {canEditURLs
              ? 'Edit ISO details and retry download. All fields can be modified for failed downloads.'
              : 'Edit ISO metadata. URLs cannot be changed for completed downloads.'}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4 mt-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Name */}
            <div>
              <label
                htmlFor="name"
                className="block text-sm font-medium mb-1.5"
              >
                Name *
              </label>
              <Input
                id="name"
                {...register('name')}
                placeholder="Ubuntu Server"
                className={errors.name ? 'border-destructive' : ''}
              />
              {errors.name && (
                <p className="text-sm text-destructive mt-1">
                  {errors.name.message}
                </p>
              )}
            </div>

            {/* Version */}
            <div>
              <label
                htmlFor="version"
                className="block text-sm font-medium mb-1.5"
              >
                Version *
              </label>
              <Input
                id="version"
                {...register('version')}
                placeholder="24.04"
                className={errors.version ? 'border-destructive' : ''}
              />
              {errors.version && (
                <p className="text-sm text-destructive mt-1">
                  {errors.version.message}
                </p>
              )}
            </div>

            {/* Architecture */}
            <div>
              <label
                htmlFor="arch"
                className="block text-sm font-medium mb-1.5"
              >
                Architecture *
              </label>
              <Select
                value={archValue}
                onValueChange={(value) => setValue('arch', value as any)}
              >
                <SelectTrigger id="arch">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="x86_64">x86_64</SelectItem>
                  <SelectItem value="aarch64">aarch64</SelectItem>
                  <SelectItem value="arm64">arm64</SelectItem>
                  <SelectItem value="i686">i686</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Edition */}
            <div>
              <label
                htmlFor="edition"
                className="block text-sm font-medium mb-1.5"
              >
                Edition
              </label>
              <Input
                id="edition"
                {...register('edition')}
                placeholder="minimal, desktop, server..."
              />
            </div>
          </div>

          {/* Download URL - only for failed ISOs */}
          <div>
            <label
              htmlFor="download_url"
              className="block text-sm font-medium mb-1.5"
            >
              Download URL *
            </label>
            <Input
              id="download_url"
              {...register('download_url')}
              placeholder="https://example.com/iso/file.iso"
              className={errors.download_url ? 'border-destructive' : ''}
              disabled={!canEditURLs}
            />
            {errors.download_url && (
              <p className="text-sm text-destructive mt-1">
                {errors.download_url.message}
              </p>
            )}
            {!canEditURLs && (
              <p className="text-sm text-muted-foreground mt-1">
                URL cannot be changed for completed downloads
              </p>
            )}
          </div>

          {/* Checksum URL - only for failed ISOs */}
          <div>
            <label
              htmlFor="checksum_url"
              className="block text-sm font-medium mb-1.5"
            >
              Checksum URL
            </label>
            <Input
              id="checksum_url"
              {...register('checksum_url')}
              placeholder="https://example.com/iso/file.iso.sha256"
              className={errors.checksum_url ? 'border-destructive' : ''}
              disabled={!canEditURLs}
            />
            {errors.checksum_url && (
              <p className="text-sm text-destructive mt-1">
                {errors.checksum_url.message}
              </p>
            )}
          </div>

          {/* Checksum Type - only for failed ISOs */}
          <div>
            <label
              htmlFor="checksum_type"
              className="block text-sm font-medium mb-1.5"
            >
              Checksum Type
            </label>
            <Select
              value={checksumTypeValue}
              onValueChange={(value) => setValue('checksum_type', value as any)}
              disabled={!canEditURLs}
            >
              <SelectTrigger id="checksum_type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sha256">SHA256</SelectItem>
                <SelectItem value="sha512">SHA512</SelectItem>
                <SelectItem value="md5">MD5</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Submit buttons */}
          <div className="flex justify-end gap-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="w-4 h-4 animate-spin" />}
              {canEditURLs ? 'Save & Retry' : 'Save Changes'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
