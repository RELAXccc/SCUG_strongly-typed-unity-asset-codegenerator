using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Unity.VisualScripting;
using UnityEngine;
using UnityEngine.EventSystems;
using UnityEngine.SceneManagement;
using UnityEngine.UI;


class MainMapUI : MonoBehaviour
{
    Button ButtonProduction;
    Button ButtonAttack;
    Button ButtonDefense;
    Button ButtonResources;
    Button ButtonSocial;

    private LoadingManager loadingManager;

    private GameObject clouds;
    private GameObject cam;
    private GameObject cloud;
    private GameObject cloud2;
    private GameObject cloud3;
    private GameObject music;

    private void Awake()
    {
        var wrapper = Prefabs.UIPrefabs.MainMap.CanvasMainMap.Get(this);
        ButtonProduction = wrapper.ImageButtonGroup.ButtonProduction.Button;
        ButtonAttack = wrapper.ImageButtonGroup.ButtonAttack.Button;
        ButtonDefense = wrapper.ImageButtonGroup.ButtonDefense.Button;
        ButtonResources = wrapper.ImageButtonGroup.ButtonResources.Button;
        ButtonSocial = wrapper.ImageButtonGroup.ButtonSocial.Button;
        loadingManager = GameObject.Find("LoadingManager").GetComponent<LoadingManager>();
        clouds = GameObject.Find("TransitionClouds");
        var cloudsWrapper = Prefabs.Loading.TransitionClouds.Get(clouds);
        cam = GameObject.Find("CameraRig");
        cloud = cloudsWrapper.Cloud.gameObject;
        cloud2 = cloudsWrapper.Cloud2.gameObject;
        cloud3 = cloudsWrapper.Cloud3.gameObject;
        music = GameObject.Find("Music");
        clouds.SetActive(false);

        if (Map.Is(Maps.Production))
        {
            ButtonProduction.image.sprite = Resources.Load<Sprite>("Images/MapButtonImages/ProductionH");
        }

        ButtonProduction.onClick.AddListener(() => MapSelect(Maps.Production));
        ButtonAttack.onClick.AddListener(() => MapSelect(Maps.Attack));
        ButtonDefense.onClick.AddListener(() => MapSelect(Maps.Defense));
        ButtonResources.onClick.AddListener(() => MapSelect(Maps.Resources));
        ButtonSocial.onClick.AddListener(() => MapSelect(Maps.Social));

        if (Game.Mode == GameMode.Town_View)
        {
            ButtonDefense.interactable = false;
            ButtonDefense.image.color = Color.grey;
        }
    }

    // smal middleware for map selection to handle special cases 
    private void MapSelect(Maps map)
    {

        if (Game.Mode == GameMode.Town_View)
        {
            LoadMap(map);
            return;
        }
        else if (Game.Mode == GameMode.Town_Manage || Game.Mode == GameMode.Defense_Edit)
        {
            UI.LastSelectedMap = map;
            SaveSystem.SavePlayer();
        }

        if (GameMode.Defense_Invasion == Game.Mode) //should never happen, but just in case
        {
            Game.ExitAttackMode();
            Map.Load(Maps.Social);
            return;
        }

        if (map == Maps.Defense)
        {
            Game.Mode = GameMode.Defense_Edit;
        }
        else
        {
            Game.Mode = GameMode.Town_Manage;
        }

        LoadMap(map);

    }

    private void LoadMap(Maps map)
    {
        clouds.SetActive(true);
        if (Map.Is(map))
        {
            UX.lib.UIHelper.UIClose();
            //Destroy(this.gameObject);
            this.gameObject.SetActive(false);
            return;
        }
        if (loadingManager != null)
        {
            // Start the loading and animation coroutine on the LoadingManager
            // Pass the scene name and all required GameObjects
            Debug.Log("Why no scene ends");
            music.GetComponent<Animator>().SetBool("SceneEnds", true);
            StartCoroutine(loadingManager.WaitAnim(
                Map.GetBuildIndex(map),
                cam,
                cloud,
                cloud2,
                cloud3,
                music,
                false
            ));

            // Optionally disable the button to prevent multiple clicks
            ButtonAttack.interactable = false;
            ButtonDefense.interactable = false;
            ButtonResources.interactable = false;
            ButtonProduction.interactable = false;
            ButtonSocial.interactable = false;

        }
        this.gameObject.GetComponent<Canvas>().enabled = false;

        //Map.Load(map);
    }

    private void OnEnable()
    {
        if (Map.Is(Maps.Production))
        {
            ButtonProduction.image.sprite = Resources.Load<Sprite>("Images/MapButtonImages/ProductionH");
        }
        else if (Map.Is(Maps.Attack))
        {
            ButtonAttack.image.sprite = Resources.Load<Sprite>("Images/MapButtonImages/AttackH");
        }
        else if (Map.Is(Maps.Defense))
        {
            ButtonDefense.image.sprite = Resources.Load<Sprite>("Images/MapButtonImages/DefenseH");
        }
        else if (Map.Is(Maps.Resources))
        {
            ButtonResources.image.sprite = Resources.Load<Sprite>("Images/MapButtonImages/ResourcesH");
        }
        else if (Map.Is(Maps.Social))
        {
            ButtonSocial.image.sprite = Resources.Load<Sprite>("Images/MapButtonImages/SocialH");
        }
    }
}

