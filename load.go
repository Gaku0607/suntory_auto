package suntory

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Gaku0607/suntory/model"
	"github.com/joho/godotenv"
)

func LoadEnvironment() error {
	if err := godotenv.Load(model.EnvPath); err != nil {
		return err
	}

	if err := loadPath(); err != nil {
		return err
	}

	if err := loadSourc(); err != nil {
		return nil
	}
	return nil
}

func loadPath() error {
	model.EnvironmentDir = os.Getenv("environment_dir")
	if model.EnvironmentDir == "" {
		return errors.New("EnvironmentDir is not exist")
	}

	model.Services_Dir = os.Getenv("services_dir")
	if model.Services_Dir == "" {
		return errors.New("ServicesDir is not exist")
	}

	model.Result_Path = os.Getenv("result_path")
	if model.Result_Path == "" {
		return errors.New("ResultPath is not exist")
	}

	return nil
}

func loadSourc() error {
	//月中庫存
	data, err := load(model.MID_MONTH_INVENTORY_BASE)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &model.Environment.MMI); err != nil {
		return err
	}

	//南部庫存管理
	data, err = load(model.SOUTH_MANAGE_TABLE_BASE)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &model.Environment.SMC); err != nil {
		return err
	}

	return nil
}

func load(filename string) ([]byte, error) {
	file, err := os.OpenFile(filepath.Join(model.EnvironmentDir, filename), os.O_RDWR|os.O_CREATE, os.ModeAppend|os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}
