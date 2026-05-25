using Prefabs.UIPrefabs.OverlayCanvas;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using UX.lib;

public class CanvasOverlay : MonoBehaviour, IUIMessageTarget
{
    private TextMeshProUGUI levelText;
    private Image experienceBarImage;
    private TextMeshProUGUI coinCounter;
    private TextMeshProUGUI xpCount;
    private Image epBarIcon;
    private TextMeshProUGUI townName;

    private Prefabs.UIPrefabs.OverlayCanvas.CanvasOverlay.Wrapper wrapper;

    //visit
    private Button ButtonExitVisitMode;
    private TextMeshProUGUI TextInfoVisit;


    private Button ButtonStock;
    private Button ButtonBuild;
    private Button ButtonChat;
    private Button ButtonQuest;
    private Button ButtonSettings;
    private Button ButtonMainMap;
    private Button ButtonLevelShield;

    private GameObject CanvasOverlayVisit;



    public static CanvasOverlay current;


    //canvases

    private GameObject canvasStock;
    private GameObject canvasBuild;
    private GameObject canvasChat;
    private GameObject canvasMainMap;
    private GameObject canvasSettings;
    private GameObject canvasQuests;
    private GameObject canvasSocial;
    private GameObject canvasDefense;

    //prefabs
    private static bool resourcesLoaded = false;
    private static GameObject canvasStockPrefab;
    private static GameObject canvasBuildPrefab;
    private static GameObject canvasChatPrefab;
    private static GameObject canvasMainMapPrefab;
    private static GameObject canvasSettingsPrefab;
    private static GameObject canvasQuestsPrefab;
    private static GameObject canvasSocialPrefab;

    private static GameObject canvasDefensePrefab = null;

    /// <summary>
    /// Intelligently retrieves the town to display.
    /// Returns the visited town if in Town_View mode, otherwise the local player.
    /// </summary>
    private Town CurrentTown
    {
        get
        {
            if (Game.Mode == GameMode.Town_View)
            {
                return Game.GetViewedTown() ?? Objects.player;
            }
            return Objects.player;
        }
    }

    private static void LoadRes()
    {
        if (resourcesLoaded) return; //Singleton
        canvasStockPrefab = Prefabs.UIPrefabs.StockMenu.CanvasStock.Load();
        canvasBuildPrefab = Prefabs.UIPrefabs.ConstructionMenu.CanvasBuild.Load();
        canvasChatPrefab = Prefabs.UIPrefabs.OverlayCanvas.ChatPopUp.ImageChatPopup.Load();
        canvasMainMapPrefab = Prefabs.UIPrefabs.MainMap.CanvasMainMap.Load();
        canvasSettingsPrefab = Prefabs.UIPrefabs.SettingsUI.CanvasSettings.Load();
        canvasQuestsPrefab = Prefabs.UIPrefabs.QuestUI.CanvasQuests.Load();
        canvasSocialPrefab = Prefabs.UIPrefabs.SocialMenu.CanvasSocial.Load();
        canvasDefensePrefab = Prefabs.UIPrefabs.DefenseMenu.CanvasDefenseBuildings.Load();
        resourcesLoaded = true;

    }

    private void Start()
    {
        current = this;
        wrapper = Prefabs.UIPrefabs.OverlayCanvas.CanvasOverlay.Get(this);

        levelText = wrapper.LevelBar.LevelbarImage.LevelbarImageShield.LevelbarText.TextMeshProUGUI;
        experienceBarImage = wrapper.LevelBar.LevelbarImage.LevelbarFrame.LevelbarMask.LevelBarProgress.Image;
        coinCounter = wrapper.CoinCounterBar.CoinCounterAmount.TextMeshProUGUI;
        xpCount = wrapper.LevelBar.LevelbarImage.LevelbarFrame.LevelbarMask.LevelBarProgress.TextXpCount.TextMeshProUGUI;
        epBarIcon = wrapper.LevelBar.LevelbarImage.LevelbarImageShield.Image;
        townName = wrapper.LevelBar.TownName.TextMeshProUGUI;

        ButtonExitVisitMode = wrapper.ButtonExitVisitmode.Button;
        TextInfoVisit = wrapper.TextInfoVisit.TextMeshProUGUI;

        ButtonExitVisitMode.onClick.AddListener(delegate
        {
            Game.ExitVisitMode();
        });

        LoadRes();

        ButtonStock = wrapper.ButtonStock.Button;
        ButtonBuild = wrapper.ButtonBuild.Button;
        ButtonChat = wrapper.ButtonChat.Button;
        ButtonQuest = wrapper.ButtonQuest.Button; //TODO: move here
        ButtonMainMap = wrapper.ButtonMainMap.Button;
        ButtonSettings = wrapper.ButtonSettings.Button;
        ButtonLevelShield = wrapper.LevelBar.LevelbarImage.LevelbarImageShield.Button;

        if (Game.Mode == GameMode.Town_View) // only when view / visit mode
        {
            ButtonStock.gameObject.SetActive(false);
            ButtonBuild.gameObject.SetActive(false);
            ButtonChat.gameObject.SetActive(false);
            ButtonQuest.gameObject.SetActive(false);

            //CanvasOverlayVisit = Instantiate(Resources.Load<GameObject>("Prefabs/UIPrefabs/OverlayCanvas/CanvasOverlayVisit"), this.transform.parent);

            // TextInfoVisit

            //Assets / Resources / Prefabs / UIPrefabs / OverlayCanvas / CanvasOverlayVisit.prefab
        }
        else // Normal Mode / View own town only
        {
            ButtonExitVisitMode.gameObject.SetActive(false);
            TextInfoVisit.gameObject.SetActive(false);
            if (canvasStock == null)
            {
                canvasStock = Instantiate(canvasStockPrefab, UIHelper.current.GetCanvases());
                var stockWrapper = Prefabs.UIPrefabs.StockMenu.CanvasStock.Get(canvasStock);
                stockWrapper.BackButtonStock.Button.onClick.AddListener(LeaveStock);
                canvasStock.SetActive(false);
            }

            //special case, Defense Edit
            if (Game.Mode == GameMode.Defense_Edit)
            {
                ButtonStock.gameObject.SetActive(false);
                ButtonBuild.onClick.AddListener(EnterDefenseBuild);
            }
            else
            {
                ButtonStock.onClick.AddListener(EnterStock);
                ButtonBuild.onClick.AddListener(EnterBuild);
            }



            ButtonChat.onClick.AddListener(OpenChat);

            ButtonSettings.onClick.AddListener(OpenSettings);

            ButtonQuest.onClick.AddListener(OpenQuests);

            ButtonLevelShield.onClick.AddListener(OpenSocial);



            //epBarIcon.GetComponent<Button>().onClick.RemoveAllListeners();
            //epBarIcon.GetComponent<Button>().onClick.AddListener(delegate
            //{
            //    OpenSocialListener();
            //});
        }

        // both modes (manage and view/visit)
        ButtonMainMap.onClick.AddListener(delegate
        {
            EnterMap();
        });


        UpdateUI();
    }

    public GameObject GetCanvasOverlayVisit()
    {
        return CanvasOverlayVisit;
    }

    private void OnEnable()
    {
        LoadRes();
        if (experienceBarImage != null) //TODO: find the true fix please
            UpdateUI();
    }



    public void UpdateUI()
    {
        if (Game.Mode == GameMode.Town_Manage && !Map.Is(Maps.Defense))
        {

            Debug.LogWarning("UIClosed called Canvas Overlay");

            Objects.LoadStock().Then(() => // remove this!!!
            {
                if (experienceBarImage == null) return;
                experienceBarImage.fillAmount = CurrentTown.progress;

                xpCount.text = CurrentTown.GetExpInfo();
                levelText.text = "" + CurrentTown.level;

                LevelIcon.SetIcon(CurrentTown.level, epBarIcon);

                Stack stack;

                if (Objects.itemStockList.TryGetValue("money", out stack))
                {
                    coinCounter.SetText(stack.amount.ToString());
                }

                townName.SetText(CurrentTown.name);
            }).Catch((ApiException api_err) =>
            {
                Logging.PrintError(api_err, Logging.DebugLevel.NecessaryLogs);
                StartCoroutine(ErrorHandler.current.SetErrorMessage(api_err.Message));
            }).Forget();

            
        }


    }

    // UI Button methods

    //Wechsle zu UI Lager

    private GameObject canvasTreeInstance;

    //LagerUI
    private GameObject contentListLager;
    public void EnterStock()
    {
        var stockWrapper = Prefabs.UIPrefabs.StockMenu.CanvasStock.Get(canvasStock);

        //test
        stockWrapper.HelpButton.Button.onClick.RemoveAllListeners();
        stockWrapper.HelpButton.Button.onClick.AddListener(delegate
        {
            if (canvasTreeInstance == null)
            {
                GameObject canvasTree = Prefabs.UIPrefabs.TreeUI.CanvasUpgradeTree.Load();
                canvasTreeInstance = Instantiate(canvasTree, UIHelper.current.GetCanvases());
                canvasTreeInstance.name = "CanvasTree Test";
            }
            else
            {
                canvasTreeInstance.SetActive(true);
            }
        });
        //

        contentListLager = stockWrapper.ImageContainerStockItems.ImageScrollStockItems.ImageListContentStockItems.gameObject;

        Objects.LoadStockGrouped().Then(stockGroups =>
        {
            foreach (var group in stockGroups)
            {
                // group.name
                if (group.stacks.Count > 0)
                {
                    GameObject container = Prefabs.UIPrefabs.StockMenu.StockItemContainer.Instantiate(contentListLager.transform);
                    var containerWrapper = Prefabs.UIPrefabs.StockMenu.StockItemContainer.Get(container);

                    containerWrapper.ImageHeadlineboxStockPrefab.TextHeadline.TextMeshProUGUI.text = group.name;

                    GameObject boxInst = containerWrapper.ImageStockItemContainerPrefab.gameObject;

                    int i = 0;
                    foreach (Stack stack in group.stacks)
                    {
                        //TODO: zahl durch spieler level erstezen
                        if (stack.item.level <= Objects.player.level)
                        {
                            // rendern
                            i++;

                            GameObject itemInst = Prefabs.UIPrefabs.StockMenu.StockItemImagePrefab.Instantiate(boxInst.transform);
                            var itemWrapper = Prefabs.UIPrefabs.StockMenu.StockItemImagePrefab.Get(itemInst);

                            itemInst.name = stack.item.id + "";

                            // image des items laden
                            itemWrapper.ImageItem.Image.sprite = stack.getSprite();
                            // menge anzeigen stack.amount
                            itemWrapper.ImageItemAmountCircle.Text.Text.text = stack.amount + "";

                            if (i == 5)
                            {
                                boxInst.GetComponent<RectTransform>().sizeDelta = new Vector2(boxInst.GetComponent<RectTransform>().sizeDelta.x, boxInst.GetComponent<RectTransform>().sizeDelta.y + 310);
                                i = 0;
                            }

                            //Button kekbutton = boxInst.transform.Find("ButtonItemStock").gameObject.GetComponent<Button>();
                            //kekbutton.onClick.AddListener(delegate { sellItemSelectionScript.SelectItem(stack); });
                        }
                    }
                }
            }
        }).Catch((ApiException api_err) =>
        {
            Logging.PrintError(api_err, Logging.DebugLevel.NecessaryLogs);
            StartCoroutine(ErrorHandler.current.SetErrorMessage(api_err.Message));
        }).Forget();


        canvasStock.SetActive(true);
        UIHelper.UIOpen();
        //loadMaterials.RefreshMaterials();


    }

    public void LeaveStock() // TODO: create a stock script
    {
        UIHelper.UIClose();
        canvasStock.SetActive(false);

        foreach (Transform child in contentListLager.transform)
        {
            GameObject.Destroy(child.gameObject);
        }


    }


    //Wechsel zu Map UI
    public void EnterMap()
    {

        if (canvasMainMap == null)
        {

            canvasMainMap = Instantiate(canvasMainMapPrefab, UIHelper.current.GetCanvases());
            canvasMainMap.name = "MainMap";

        }
        else
        {
            canvasMainMap.SetActive(true);
        }


        UIHelper.UIOpen();
    }

    //Wechsel UI Build
    public void EnterBuild()
    {
        if (canvasBuild == null)
        {
            canvasBuild = Instantiate(canvasBuildPrefab, UIHelper.current.GetCanvases());
            canvasBuild.name = "CanvasBuild";
        }
        else
        {
            canvasBuild.SetActive(true);
        }

        UIHelper.UIOpen();
    }


    //Wechsel UI Defense Build
    public void EnterDefenseBuild()
    {


        if (canvasDefense == null)
        {
            canvasDefense = Instantiate(canvasDefensePrefab, UIHelper.current.GetCanvases());
            canvasDefense.name = "CanvasDefense";
        }
        else
        {
            canvasDefense.SetActive(true);
        }
        UIHelper.UIOpen();
    }



    public void OpenChat()
    {

        if (canvasChat == null)
        {
            canvasChat = Instantiate(canvasChatPrefab, this.transform);
            canvasChat.name = "ChatPopup";
        }
        else
        {
            canvasChat.SetActive(true);
        }

        var rectTransform = canvasChat.GetComponent<RectTransform>().anchoredPosition = new Vector2(0, 0);

        //UIHelper.current.AddTempPopup(canvasChat);
    }

    public void CloseChat()
    {
        UIHelper.current.CloseAllTempPopups();
    }

    private void OpenSettings()
    {

        if (canvasSettings == null)
        {
            canvasSettings = Instantiate(canvasSettingsPrefab, UIHelper.current.GetCanvases());
            canvasSettings.name = "CanvasSettings";
        }
        else
        {
            canvasSettings.SetActive(true);
        }
        UIHelper.UIOpen();
    }
    private void OpenQuests()
    {
        Objects.LoadItemStock().Forget();

        if (canvasQuests == null)
        {
            canvasQuests = Instantiate(canvasQuestsPrefab, UIHelper.current.GetCanvases());
            canvasQuests.name = "CanvasQuests";
        }
        else
        {
            canvasQuests.SetActive(true);
        }
        UIHelper.UIOpen();
    }

    private void OpenSocial()
    {

        if (canvasSocial == null)
        {
            canvasSocial = Instantiate(canvasSocialPrefab, UIHelper.current.GetCanvases());
            canvasSocial.name = "CanvasSocial";
        }
        else
        {
            canvasSocial.SetActive(true);
        }
        UIHelper.UIOpen();
    }
}
