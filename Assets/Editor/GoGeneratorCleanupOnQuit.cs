using UnityEngine;
using UnityEditor;
using System.IO;

/// <summary>
/// Cleans up temporary files (executable and cache) when Unity is closed.
/// This ensures the project stays clean and avoids binary bloat in source control.
/// </summary>
[InitializeOnLoad]
public class GoGeneratorCleanupOnQuit
{
    static GoGeneratorCleanupOnQuit()
    {
        EditorApplication.quitting += Cleanup;
    }

    private static void Cleanup()
    {
        string projectRoot = Directory.GetParent(Application.dataPath).FullName;
        string generatorPath = Path.Combine(projectRoot, "generator");

        if (!Directory.Exists(generatorPath)) return;

        // 1. Delete the compiled executable
        string exeName = Application.platform == RuntimePlatform.WindowsEditor ? "scug.exe" : "scug";
        string exePath = Path.Combine(generatorPath, exeName);
        
        try
        {
            if (File.Exists(exePath))
            {
                File.Delete(exePath);
                // We use Console.WriteLine or similar if needed, but Unity's Debug.Log 
                // might not show up reliably during the quitting phase.
            }
        }
        catch (System.Exception e)
        {
            // Silent fail on quit is usually preferred, but we can log to a file if needed
        }

        // 2. Delete the cache file
        string cachePath = Path.Combine(generatorPath, "scug_cache.json");
        try
        {
            if (File.Exists(cachePath))
            {
                File.Delete(cachePath);
            }
        }
        catch (System.Exception)
        {
            // Ignore
        }
    }
}
