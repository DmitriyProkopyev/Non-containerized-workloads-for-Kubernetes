import React from 'react';

export default function StandardSettings({ settings, setSettings }) {
  const handleChange = (field, value) => {
    setSettings(prev => ({ ...prev, [field]: value }));
  };

  const handleNestedChange = (field, value) => {
    setSettings(prev => ({
      ...prev,
      settings: { ...prev.settings, [field]: value }
    }));
  };

  const handleNotificationChange = (e) => {
    setSettings(prev => ({
      ...prev,
      settings: {
        ...prev.settings,
        notifications: { email: e.target.checked }
      }
    }));
  };

  return (
    <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
      <h2 className="text-xl font-semibold mb-4">Standard Settings</h2>
      
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">User Name</label>
          <input
            className="w-full p-2 border rounded-md"
            value={settings.username}
            onChange={(e) => handleChange('username', e.target.value)}
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Email</label>
          <input
            className="w-full p-2 border rounded-md"
            value={settings.email}
            onChange={(e) => handleChange('email', e.target.value)}
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Language</label>
          <select
            className="w-full p-2 border rounded-md"
            value={settings.settings.language}
            onChange={(e) => handleNestedChange('language', e.target.value)}
          >
            <option value="English">English</option>
            <option value="Russian">Russian</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Theme</label>
          <select
            className="w-full p-2 border rounded-md"
            value={settings.settings.theme}
            onChange={(e) => handleNestedChange('theme', e.target.value)}
          >
            <option value="Light">Light</option>
            <option value="Dark">Dark</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={settings.settings.notifications.email}
            onChange={handleNotificationChange}
            className="w-4 h-4"
          />
          <label className="text-sm">Email notifications</label>
        </div>
      </div>
    </div>
  );
}
