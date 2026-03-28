'use client';

import { useState, useRef, useCallback } from 'react';
import { Upload, X, FileText, CheckCircle } from 'lucide-react';
import { cn } from '@/lib/utils';

// ── Types ────────────────────────────────────────────────
interface FileUploadProps {
  onUpload: (file: File) => void | Promise<void>;
  accept?: string;
  maxSize?: number; // in bytes
  label?: string;
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// ── Component ────────────────────────────────────────────
export function FileUpload({
  onUpload,
  accept,
  maxSize,
  label = 'Upload a file',
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragActive, setDragActive] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [uploaded, setUploaded] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const validateAndSet = useCallback(
    (file: File) => {
      setError(null);
      setUploaded(false);

      if (maxSize && file.size > maxSize) {
        setError(`File exceeds maximum size of ${formatFileSize(maxSize)}.`);
        return;
      }

      if (accept) {
        const allowed = accept.split(',').map((t) => t.trim());
        const ext = '.' + file.name.split('.').pop()?.toLowerCase();
        const typeMatch = allowed.some(
          (a) => a === file.type || a === ext || (a.endsWith('/*') && file.type.startsWith(a.replace('/*', '/'))),
        );
        if (!typeMatch) {
          setError(`File type not allowed. Accepted: ${accept}`);
          return;
        }
      }

      setSelectedFile(file);
    },
    [accept, maxSize],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragActive(false);

      const file = e.dataTransfer.files?.[0];
      if (file) validateAndSet(file);
    },
    [validateAndSet],
  );

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) validateAndSet(file);
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setUploading(true);
    setError(null);
    try {
      await onUpload(selectedFile);
      setUploaded(true);
    } catch {
      setError('Upload failed. Please try again.');
    } finally {
      setUploading(false);
    }
  };

  const handleClear = () => {
    setSelectedFile(null);
    setUploaded(false);
    setError(null);
    if (inputRef.current) inputRef.current.value = '';
  };

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-gray-700">
        {label}
      </label>

      {/* Drop Zone */}
      {!selectedFile && (
        <div
          onDragOver={(e) => {
            e.preventDefault();
            setDragActive(true);
          }}
          onDragLeave={() => setDragActive(false)}
          onDrop={handleDrop}
          onClick={() => inputRef.current?.click()}
          className={cn(
            'flex cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed px-6 py-10 text-center transition-colors',
            dragActive
              ? 'border-indigo-400 bg-indigo-50'
              : 'border-gray-300 bg-white hover:border-gray-400 hover:bg-gray-50',
          )}
        >
          <Upload
            className={cn(
              'h-8 w-8 mb-3',
              dragActive ? 'text-indigo-500' : 'text-gray-400',
            )}
          />
          <p className="text-sm text-gray-600">
            <span className="font-semibold text-indigo-600">Click to browse</span> or drag and
            drop
          </p>
          {accept && (
            <p className="mt-1 text-xs text-gray-400">
              Accepted: {accept}
            </p>
          )}
          {maxSize && (
            <p className="mt-0.5 text-xs text-gray-400">
              Max size: {formatFileSize(maxSize)}
            </p>
          )}

          <input
            ref={inputRef}
            type="file"
            accept={accept}
            onChange={handleFileChange}
            className="hidden"
          />
        </div>
      )}

      {/* Selected File */}
      {selectedFile && (
        <div className="flex items-center gap-3 rounded-lg border border-gray-200 bg-white px-4 py-3">
          <FileText className="h-5 w-5 shrink-0 text-gray-400" />
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-gray-900">
              {selectedFile.name}
            </p>
            <p className="text-xs text-gray-500">
              {formatFileSize(selectedFile.size)}
            </p>
          </div>

          {uploaded && (
            <CheckCircle className="h-5 w-5 shrink-0 text-green-500" />
          )}

          {!uploading && (
            <button
              onClick={handleClear}
              className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
              aria-label="Remove file"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>
      )}

      {/* Upload Progress / Button */}
      {selectedFile && !uploaded && (
        <button
          onClick={handleUpload}
          disabled={uploading}
          className={cn(
            'btn-primary w-full',
            uploading && 'cursor-not-allowed opacity-70',
          )}
        >
          {uploading ? (
            <span className="flex items-center gap-2">
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              Uploading...
            </span>
          ) : (
            'Upload File'
          )}
        </button>
      )}

      {/* Error */}
      {error && (
        <p className="text-xs text-red-600" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
