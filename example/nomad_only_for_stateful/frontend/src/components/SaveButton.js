import React, { useState, useEffect, useRef } from 'react';
import { FiSave, FiCheck, FiEdit } from 'react-icons/fi';

export default function SaveButton({ settings, documentId, onSaveSuccess }) {
  const [saved, setSaved] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const initialSettings = useRef(null);

  useEffect(() => {
    if (documentId && !initialSettings.current) {
      initialSettings.current = JSON.parse(JSON.stringify(settings));
      setSaved(true); 
    }
  }, [documentId, settings]);

  useEffect(() => {
    if (!initialSettings.current) return;
    
    const isEqual = JSON.stringify(settings) === JSON.stringify(initialSettings.current);
    setHasChanges(!isEqual);
    
    if (!isEqual) setSaved(false);
  }, [settings]);

  const handleSave = async () => {
    try {
      const method = documentId ? 'PUT' : 'POST';
      const url = documentId ? `/update_document/${documentId}` : '/create_document/';

      const { _id, user_id, ...dataToSend } = settings;

      const response = await fetch(url, {
        method,
        headers: { 
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({
          ...dataToSend,
          ...(!documentId && { user_id: settings.user_id })
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.detail || 'Save failed');
      }

      const data = await response.json();
      onSaveSuccess(data);
      setSaved(true);
      
      initialSettings.current = JSON.parse(JSON.stringify(settings));
      setHasChanges(false);
    } catch (error) {
      console.error('Save error:', error);
      alert(`Error: ${error.message}`);
    }
  };

  return (
    <button
      onClick={handleSave}
      className={`flex items-center gap-2 px-6 py-3 rounded-lg transition-all duration-300
        ${saved 
          ? 'bg-green-600 hover:bg-green-700 text-white' 
          : 'bg-blue-600 hover:bg-blue-700 text-white'
        }`}
      disabled={saved}
    >
      {saved ? (
        <>
          <FiCheck className="w-5 h-5" />
          Saved!
        </>
      ) : (
        <>
          {documentId ? <FiEdit className="w-5 h-5" /> : <FiSave className="w-5 h-5" />}
          {documentId ? 'Update Settings' : 'Save Settings'}
        </>
      )}
    </button>
  );
}