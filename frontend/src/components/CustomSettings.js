import React from 'react';
import { FiTrash2, FiPlus } from 'react-icons/fi';

export default function CustomSettings({ settings, setSettings }) {
  const customSettings = settings.settings.custom || {};

  const handleAdd = () => {
    const newKey = prompt("Enter setting name:");
    if (newKey && !(newKey in customSettings)) {
      setSettings(prev => ({
        ...prev,
        settings: {
          ...prev.settings,
          custom: { ...prev.settings.custom, [newKey]: "" }
        }
      }));
    }
  };

  const handleDelete = (key) => {
    const updatedCustom = { ...customSettings };
    delete updatedCustom[key];
    setSettings(prev => ({
      ...prev,
      settings: { ...prev.settings, custom: updatedCustom }
    }));
  };

  const handleValueChange = (key, value) => {
    setSettings(prev => ({
      ...prev,
      settings: {
        ...prev.settings,
        custom: { ...prev.settings.custom, [key]: value }
      }
    }));
  };

  const handleTypeChange = (key, type) => {
    let newValue;
    switch (type) {
      case 'string':
        newValue = '';
        break;
      case 'number':
        newValue = 0;
        break;
      case 'boolean':
        newValue = false;
        break;
      default:
        newValue = '';
    }
    setSettings(prev => ({
      ...prev,
      settings: {
        ...prev.settings,
        custom: { ...prev.settings.custom, [key]: newValue }
      }
    }));
  };

  return (
    <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">Custom Settings</h2>
        <button
          onClick={handleAdd}
          className="flex items-center gap-1 px-3 py-1.5 text-sm bg-blue-100 text-blue-600 rounded-md hover:bg-blue-200"
        >
          <FiPlus className="w-4 h-4" />
          Add
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left p-2 text-sm font-medium">Key</th>
              <th className="text-left p-2 text-sm font-medium">Value</th>
              <th className="text-left p-2 text-sm font-medium">Type</th>
              <th className="text-left p-2 text-sm font-medium">Delete</th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(customSettings).map(([key, value]) => (
              <tr key={key} className="border-t h-14">
                <td className="p-2">{key}</td>
                <td className="p-2">
                  <input
                    className="w-full p-1.5 border rounded-md"
                    value={value}
                    onChange={(e) => handleValueChange(key, e.target.value)}
                  />
                </td>
                <td className="p-2">
                  <select
                    className="w-full p-1.5 border rounded-md"
                    value={typeof value}
                    onChange={(e) => handleTypeChange(key, e.target.value)}
                  >
                    <option value="string">string</option>
                    <option value="number">number</option>
                    <option value="boolean">boolean</option>
                  </select>
                </td>
                <td className="p-2 text-center">
                  <button
                    onClick={() => handleDelete(key)}
                    className="text-red-500 hover:text-red-600"
                  >
                    <FiTrash2 className="w-5 h-5" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
