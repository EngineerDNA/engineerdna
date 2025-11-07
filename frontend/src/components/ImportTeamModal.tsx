import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import Papa from 'papaparse';
import { api } from '../api/client';

interface ImportTeamModalProps {
  onClose: () => void;
}

const KNOWN_COLUMNS = [
  'name',
  'email',
  'manager',
  'github',
  'jira',
  'gitlab',
  'slack',
  'zoom',
  'csv',
];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

export function ImportTeamModal({ onClose }: ImportTeamModalProps) {
  const queryClient = useQueryClient();
  const [file, setFile] = useState<File | null>(null);
  const [headers, setHeaders] = useState<string[]>([]);
  const [preview, setPreview] = useState<string[][]>([]);
  const [columnMapping, setColumnMapping] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => {
      if (!file) {
        throw new Error('No file selected');
      }
      return api.importEngineers(file, columnMapping);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['engineers'] });
      onClose();
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (!selectedFile) return;

    // Check file size before processing
    if (selectedFile.size > MAX_FILE_SIZE) {
      setError(`File too large (max ${MAX_FILE_SIZE / 1024 / 1024}MB)`);
      return;
    }

    setFile(selectedFile);
    setError(null);

    try {
      const text = await selectedFile.text();

      // Use papaparse for proper CSV parsing
      Papa.parse(text, {
        header: true,
        skipEmptyLines: true,
        transformHeader: (header: string) => header.trim().toLowerCase(),
        complete: (results) => {
          if (results.errors.length > 0) {
            setError(`CSV parsing error: ${results.errors[0].message}`);
            return;
          }

          if (!results.data || results.data.length === 0) {
            setError('File is empty');
            return;
          }

          const csvHeaders = results.meta.fields || [];
          setHeaders(csvHeaders);

          // Preview first 5 rows
          const previewData = results.data.slice(0, 5) as Record<string, string>[];
          const previewRows = previewData.map((row) =>
            csvHeaders.map((header) => row[header] || '')
          );
          setPreview(previewRows);

          // Auto-map known columns
          const autoMapping: Record<string, string> = {};
          KNOWN_COLUMNS.forEach((field) => {
            const matchingHeader = csvHeaders.find((h) => h.toLowerCase() === field.toLowerCase());
            if (matchingHeader) {
              autoMapping[field] = matchingHeader;
            }
          });
          setColumnMapping(autoMapping);
        },
        error: (error: Error) => {
          setError(`Failed to parse CSV file: ${error.message}`);
        },
      });
    } catch (err) {
      setError('Failed to read CSV file');
    }
  };

  const handleMappingChange = (targetField: string, csvColumn: string) => {
    setColumnMapping((prev) => {
      if (csvColumn === '') {
        // Remove mapping if user selects "Skip this column"
        const newMapping = { ...prev };
        delete newMapping[targetField];
        return newMapping;
      }
      return {
        ...prev,
        [targetField]: csvColumn,
      };
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!file) {
      setError('Please select a file');
      return;
    }

    if (!columnMapping.name) {
      setError('Name column mapping is required');
      return;
    }

    mutation.mutate();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <h2 className="text-2xl font-bold mb-4">Import Team from CSV</h2>

        <form onSubmit={handleSubmit}>
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Select CSV File
              </label>
              <input
                type="file"
                accept=".csv"
                onChange={handleFileChange}
                className="block w-full text-sm text-gray-900 border border-gray-300 rounded cursor-pointer bg-gray-50 focus:outline-none"
              />
              <p className="mt-1 text-sm text-gray-500">
                CSV file with columns: name, email, manager, github, jira, etc.
              </p>
            </div>

            {headers.length > 0 && (
              <>
                <div>
                  <h3 className="text-lg font-medium mb-3">Column Mapping</h3>
                  <div className="space-y-2">
                    {KNOWN_COLUMNS.map((targetField) => (
                      <div key={targetField} className="flex items-center gap-4">
                        <div className="w-1/3 text-sm font-medium text-gray-700">
                          Target Field: <span className="text-blue-600">{targetField}</span>
                        </div>
                        <div className="w-1/3">
                          <select
                            value={columnMapping[targetField] || ''}
                            onChange={(e) => handleMappingChange(targetField, e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                          >
                            <option value="">Skip this field</option>
                            {headers.map((header) => (
                              <option key={header} value={header}>
                                {header}
                              </option>
                            ))}
                          </select>
                        </div>
                        <div className="w-1/3 text-sm text-gray-600">
                          {targetField === 'name' && <span className="text-red-600">Required</span>}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-lg font-medium mb-3">Preview (First 5 Rows)</h3>
                  <div className="overflow-x-auto">
                    <table className="min-w-full border border-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          {headers.map((header) => (
                            <th
                              key={header}
                              className="px-4 py-2 text-left text-xs font-medium text-gray-700 uppercase"
                            >
                              {header}
                              {Object.entries(columnMapping).some(
                                ([, csvCol]) => csvCol === header
                              ) && (
                                <div className="text-blue-600 normal-case">
                                  →{' '}
                                  {
                                    Object.entries(columnMapping).find(
                                      ([, csvCol]) => csvCol === header
                                    )?.[0]
                                  }
                                </div>
                              )}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {preview.map((row, rowIndex) => (
                          <tr key={rowIndex} className="border-t">
                            {row.map((cell, cellIndex) => (
                              <td key={cellIndex} className="px-4 py-2 text-sm text-gray-900">
                                {cell}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            )}
          </div>

          {error && (
            <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded text-red-800">
              {error}
            </div>
          )}

          <div className="mt-6 flex gap-3 justify-end">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 border border-gray-300 rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={mutation.isPending || !file}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {mutation.isPending ? 'Importing...' : 'Import'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
