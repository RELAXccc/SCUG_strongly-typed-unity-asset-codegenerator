using UnityEngine;
using UnityEditor;
using System.Diagnostics;
using System.IO;

/// <summary>
/// Ensures the Go generator is compiled and ready to use when Unity opens.
/// If no cache exists, it triggers a full scan to prime the system.
/// </summary>
[InitializeOnLoad]
public class GoGeneratorBuildOnStartup
{
    static GoGeneratorBuildOnStartup()
    {
        BuildGenerator();
    }

    public static void BuildGenerator()
    {
        string projectRoot = Directory.GetParent(Application.dataPath).FullName;
        string generatorPath = Path.Combine(projectRoot, "generator");

        if (!Directory.Exists(generatorPath)) return;

        string exeName = Application.platform == RuntimePlatform.WindowsEditor ? "scug.exe" : "scug";
        string exePath = Path.Combine(generatorPath, exeName);

        UnityEngine.Debug.Log($"[Go Generator] Checking/Building generator executable...");

        ProcessStartInfo startInfo = new ProcessStartInfo();
        startInfo.WorkingDirectory = generatorPath;

        if (Application.platform == RuntimePlatform.WindowsEditor)
        {
            startInfo.FileName = "cmd.exe";
            startInfo.Arguments = $"/c go build -o {exeName} cmd/scug/main.go";
        }
        else
        {
            startInfo.FileName = "go";
            startInfo.Arguments = $"build -o {exeName} cmd/scug/main.go";
        }

        startInfo.UseShellExecute = false;
        startInfo.CreateNoWindow = true;
        startInfo.RedirectStandardError = true;

        try
        {
            using (Process process = Process.Start(startInfo))
            {
                string error = process.StandardError.ReadToEnd();
                process.WaitForExit();

                if (process.ExitCode == 0)
                {
                    UnityEngine.Debug.Log($"[Go Generator] Build successful: {exeName}");
                    
                    // NEW: Check if cache is missing. If so, prime it in the background.
                    string cachePath = Path.Combine(generatorPath, "scug_cache.json");
                    if (!File.Exists(cachePath))
                    {
                        UnityEngine.Debug.Log("[Go Generator] Cache missing. Starting background prime (full scan)...");
                        // Pass task name "Full Scan" for better logs
                        GoGeneratorOnSave.RunGoGenerator(null, "Full Scan");
                    }
                }
                else
                {
                    UnityEngine.Debug.LogError($"[Go Generator] Build FAILED:\n{error}");
                }
            }
        }
        catch (System.Exception e)
        {
            UnityEngine.Debug.LogError($"[Go Generator] Build Exception: {e.Message}");
        }
    }
}
