import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import type { CreateISORequest } from '../types/iso';
import { Plus, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';

const createISOSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  version: z.string().min(1, 'Version is required'),
  arch: z.enum(['x86_64', 'aarch64', 'arm64', 'i686']),
  edition: z.string().optional(),
  download_url: z.string().url('Must be a valid URL'),
  checksum_url: z.string().url('Must be a valid URL').optional().or(z.literal('')),
  checksum_type: z.enum(['sha256', 'sha512', 'md5']).optional(),
});

type CreateISOFormData = z.infer<typeof createISOSchema>;

interface AddIsoFormProps {
  onSubmit: (request: CreateISORequest) => Promise<void>;
}

export function AddIsoForm({ onSubmit }: AddIsoFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
    watch,
    setValue,
  } = useForm<CreateISOFormData>({
    resolver: zodResolver(createISOSchema),
    defaultValues: {
      name: '',
      version: '',
      arch: 'x86_64',
      edition: '',
      download_url: '',
      checksum_url: '',
      checksum_type: 'sha256',
    },
  });

  const archValue = watch('arch');
  const checksumTypeValue = watch('checksum_type');

  const onFormSubmit = async (data: CreateISOFormData) => {
    try {
      await onSubmit(data as CreateISORequest);
      reset();
      // Close dialog by resetting form - dialog is controlled by open state
    } catch (error) {
      // Error handling is done in the parent component
      console.error('Failed to create ISO:', error);
    }
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>
          <Plus />
          Add ISO Download
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add New ISO Download</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4 mt-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium mb-1.5">
                Name *
              </label>
              <Input
                id="name"
                {...register('name')}
                placeholder="Alpine Linux"
                aria-invalid={!!errors.name}
              />
              {errors.name && <p className="text-xs text-destructive mt-1">{errors.name.message}</p>}
            </div>

            <div>
              <label htmlFor="version" className="block text-sm font-medium mb-1.5">
                Version *
              </label>
              <Input
                id="version"
                {...register('version')}
                placeholder="3.19.1"
                aria-invalid={!!errors.version}
              />
              {errors.version && <p className="text-xs text-destructive mt-1">{errors.version.message}</p>}
            </div>

            <div>
              <label htmlFor="arch" className="block text-sm font-medium mb-1.5">
                Architecture *
              </label>
              <Select value={archValue} onValueChange={(value) => setValue('arch', value as any)}>
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
              {errors.arch && <p className="text-xs text-destructive mt-1">{errors.arch.message}</p>}
            </div>

            <div>
              <label htmlFor="edition" className="block text-sm font-medium mb-1.5">
                Edition
              </label>
              <Input id="edition" {...register('edition')} placeholder="minimal, desktop, server" />
            </div>
          </div>

          <div>
            <label htmlFor="download_url" className="block text-sm font-medium mb-1.5">
              Download URL *
            </label>
            <Input
              id="download_url"
              {...register('download_url')}
              placeholder="https://example.com/alpine-3.19.1-x86_64.iso"
              className="font-mono text-sm"
              aria-invalid={!!errors.download_url}
            />
            {errors.download_url && (
              <p className="text-xs text-destructive mt-1">{errors.download_url.message}</p>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="md:col-span-2">
              <label htmlFor="checksum_url" className="block text-sm font-medium mb-1.5">
                Checksum URL
              </label>
              <Input
                id="checksum_url"
                {...register('checksum_url')}
                placeholder="https://example.com/alpine-3.19.1-x86_64.iso.sha256"
                className="font-mono text-sm"
                aria-invalid={!!errors.checksum_url}
              />
              {errors.checksum_url && (
                <p className="text-xs text-destructive mt-1">{errors.checksum_url.message}</p>
              )}
            </div>

            <div>
              <label htmlFor="checksum_type" className="block text-sm font-medium mb-1.5">
                Checksum Type
              </label>
              <Select
                value={checksumTypeValue}
                onValueChange={(value) => setValue('checksum_type', value as any)}
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
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus />
                  Create Download
                </>
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
