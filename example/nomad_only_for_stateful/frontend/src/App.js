import { useState, useEffect } from 'react';
import StandardSettings from './components/StandardSettings';
import CustomSettings from './components/CustomSettings';
import SaveButton from './components/SaveButton';

export default function App() {
  const [settings, setSettings] = useState({
    user_id: "abc123",
    username: "",
    email: "",
    settings: {
      language: "",
      theme: "",
      notifications: { email: false },
      custom: {}
    },
    updated_at: ""
  });
  const [documentId, setDocumentId] = useState(localStorage.getItem('lastDocumentId') || "");
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadSettings = async () => {
      try {
        if (!documentId) {
          setIsLoading(false);
          return;
        }

        const response = await fetch(`/read_document/${documentId}`);
        if (!response.ok) throw new Error('Failed to load settings');
        
        const data = await response.json();
        setSettings(data);
      } catch (error) {
        console.error('Load error:', error);
        localStorage.removeItem('lastDocumentId');
      } finally {
        setIsLoading(false);
      }
    };

    loadSettings();
  }, []);

  const handleSaveSuccess = (savedDocument) => {
    const newId = savedDocument._id?.$oid || "";
    localStorage.setItem('lastDocumentId', newId);
    setDocumentId(newId);
  };

  if (isLoading) {
    return (
      <div className="container mx-auto p-6 text-center">
        <div className="animate-spin inline-block w-8 h-8 border-4 rounded-full border-t-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <h1 className="text-3xl font-bold text-center">Settings Panel</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <StandardSettings settings={settings} setSettings={setSettings} />
        <CustomSettings settings={settings} setSettings={setSettings} />
      </div>

      <div className="flex justify-center">
        <SaveButton 
          settings={settings} 
          documentId={documentId}
          onSaveSuccess={handleSaveSuccess} 
        />
      </div>
    </div>
  );
}