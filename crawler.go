package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

type crawler struct {
	collectionTimeout time.Duration
	timeBetweenSteps  time.Duration
	year              string
	month             string
	output            string
}

const (
	contrachequeXPATH = "//*[@id='16']/div[3]/table/tbody/tr/td"
	indenizacoesXPATH = "//*[@id='67']"
)

func (c crawler) crawl() ([]string, error) {
	// Chromedp setup.
	log.SetOutput(os.Stderr) // Enviando logs para o stderr para não afetar a execução do coletor.
	alloc, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
			chromedp.Flag("headless", false), // mude para false para executar com navegador visível.
			chromedp.NoSandbox,
			chromedp.DisableGPU,
		)...,
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(
		alloc,
		chromedp.WithLogf(log.Printf), // remover comentário para depurar
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, c.collectionTimeout)
	defer cancel()

	// NOTA IMPORTANTE: os prefixos dos nomes dos arquivos tem que ser igual
	// ao esperado no parser MPSC.

	// Seleciona o contracheque na página principal
	log.Printf("Clicando em contracheque(%s/%s)...", c.month, c.year)
	if err := c.navegacaoSite(ctx, contrachequeXPATH); err != nil {
		log.Fatalf("Erro no setup:%v", err)
	}
	log.Printf("Clicado com sucesso!\n")

	// Contracheque
	log.Printf("Realizando seleção (%s/%s)...", c.month, c.year)
	if err := c.selecionaMesAno(ctx); err != nil {
		log.Fatalf("Erro no setup:%v", err)
	}
	log.Printf("Seleção realizada com sucesso!\n")
	cqFname := c.downloadFilePath("contracheque")
	log.Printf("Fazendo download do contracheque (%s)...", cqFname)

	if err := c.exportaPlanilha(ctx, cqFname, "contra"); err != nil {
		log.Fatalf("Erro fazendo download do contracheque: %v", err)
	}
	log.Printf("Download realizado com sucesso!\n")

	// Indenizações
	log.Printf("\nClicando na aba indenizações (%s/%s)...", c.month, c.year)
	if err := c.clicaAba(ctx, indenizacoesXPATH); err != nil {
		log.Fatalf("Erro no setup:%v", err)
	}

	log.Printf("Realizando seleção (%s/%s)...", c.month, c.year)
	if err := c.selecionaMesAnoVerbas(ctx); err != nil {
		log.Fatalf("Erro no setup:%v", err)
	}

	// log.Printf("Seleção realizada com sucesso!\n")
	iFname := c.downloadFilePath("verbas-indenizatorias")
	log.Printf("Fazendo download das indenizações (%s)...", iFname)
	if err := c.exportaPlanilha(ctx, iFname, "verbas"); err != nil {
		log.Fatalf("Erro fazendo download dos indenizações: %v", err)
	}
	log.Printf("Download realizado com sucesso!\n")

	// Retorna caminhos completos dos arquivos baixados.
	return []string{cqFname, iFname}, nil
}

func (c crawler) downloadFilePath(prefix string) string {
	return filepath.Join(c.output, fmt.Sprintf("membros-ativos-%s-%s-%s.xlsx", prefix, c.month, c.year))
}

// Navega para as planilhas
func (c crawler) navegacaoSite(ctx context.Context, xpath string) error {
	const (
		baseURL = "https://transparencia.mpsc.mp.br/QvAJAXZfc/opendoc.htm?document=Portal%20Transparencia%2FPortal%20Transp%20MPSC.qvw&host=QVS%40qvias&anonymous=false"
	)

	return chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.Sleep(c.timeBetweenSteps),

		// Abre o contracheque
		chromedp.Click(xpath, chromedp.BySearch, chromedp.NodeReady),
		chromedp.Sleep(c.timeBetweenSteps),

		// Altera o diretório de download
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(c.output).
			WithEventsEnabled(true),
	)
}

// clicaAba clica na aba referenciada pelo XPATH passado como parâmetro.
// Também espera até o estar visível.
func (c crawler) clicaAba(ctx context.Context, xpath string) error {
	return chromedp.Run(ctx,
		chromedp.Click(xpath),
		chromedp.Sleep(c.timeBetweenSteps),
	)
}

func (c crawler) selecionaMesAno(ctx context.Context) error {
	if c.month == "01" {
		c.month = "jan"
	} else if c.month == "02" {
		c.month = "fev"
	} else if c.month == "03" {
		c.month = "mar"
	} else if c.month == "04" {
		c.month = "abr"
	} else if c.month == "05" {
		c.month = "mai"
	} else if c.month == "06" {
		c.month = "jun"
	} else if c.month == "07" {
		c.month = "jul"
	} else if c.month == "08" {
		c.month = "ago"
	} else if c.month == "09" {
		c.month = "set"
	} else if c.month == "10" {
		c.month = "out"
	} else if c.month == "11" {
		c.month = "nov"
	} else if c.month == "12" {
		c.month = "dez"
	}

	month := fmt.Sprintf("//*[@title='%s']", c.month)
	year := fmt.Sprintf("//*[@title='%s']", c.year)
	if c.year != "2021" {
		if c.month != "dez" {
			return chromedp.Run(ctx,
				// Espera ficar visível
				chromedp.WaitVisible(year, chromedp.BySearch, chromedp.NodeReady),
				chromedp.Sleep(c.timeBetweenSteps),

				// Seleciona o ano
				chromedp.Click(year, chromedp.BySearch, chromedp.NodeReady),
				chromedp.Sleep(c.timeBetweenSteps),

				// Seleciona o mes
				chromedp.Click(month, chromedp.BySearch, chromedp.NodeVisible),
				chromedp.Sleep(c.timeBetweenSteps),
			)
		}
		return chromedp.Run(ctx,
			// Espera ficar visível
			chromedp.WaitVisible(year, chromedp.BySearch, chromedp.NodeReady),
			chromedp.Sleep(c.timeBetweenSteps),

			// Seleciona o ano
			chromedp.Click(year, chromedp.BySearch, chromedp.NodeReady),
			chromedp.Sleep(c.timeBetweenSteps),
		)
	}

	return chromedp.Run(ctx)
}
	

func (c crawler) selecionaMesAnoVerbas(ctx context.Context) error {
	var pathMonth string
	var pathYear string

	if c.year == "2018"{
		pathYear = "/html/body/div[5]/div/div[58]/div[3]/div/div[1]/div[3]"
	} else if c.year == "2019"{
		pathYear = "/html/body/div[5]/div/div[67]/div[3]/div/div[1]/div[4]"
	} else if c.year == "2020"{
		pathYear = "/html/body/div[5]/div/div[67]/div[3]/div/div[1]/div[5]"
	} else if c.year == "2021"{
		pathYear = "/html/body/div[5]/div/div[67]/div[3]/div/div[1]/div[6]"
	} else if c.year == "2022"{
		pathYear = "/html/body/div[5]/div/div[67]/div[3]/div/div[1]/div[7]"
	}

	if c.month == "01" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[1]"
	} else if c.month == "02" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[2]"
	} else if c.month == "03" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[3]"
	} else if c.month == "04" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[4]"
	} else if c.month == "05" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[5]"
	} else if c.month == "06" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[6]"
	} else if c.month == "07" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[7]"
	} else if c.month == "08" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[8]"
	} else if c.month == "09" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[9]"
	} else if c.month == "10" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[10]"
	} else if c.month == "11" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[11]"
	} else if c.month == "12" {
		pathMonth = "/html/body/div[5]/div/div[64]/div[3]/div/div[1]/div[12]"
	}
	
	if c.year != "2021" {
		if c.month != "dez" {
			return chromedp.Run(ctx,
				// Espera ficar visível
				chromedp.WaitVisible(pathYear, chromedp.BySearch, chromedp.NodeReady),
				chromedp.Sleep(c.timeBetweenSteps),

				// Seleciona o ano
				chromedp.Click(pathYear, chromedp.BySearch, chromedp.NodeReady),
				chromedp.Sleep(c.timeBetweenSteps),

				// Seleciona o mes
				chromedp.Click(pathMonth, chromedp.BySearch, chromedp.NodeVisible),
				chromedp.Sleep(c.timeBetweenSteps),
			)
		}
		return chromedp.Run(ctx,
			// Espera ficar visível
			chromedp.WaitVisible(pathYear, chromedp.BySearch, chromedp.NodeReady),
			chromedp.Sleep(c.timeBetweenSteps),

			// Seleciona o ano
			chromedp.Click(pathYear, chromedp.BySearch, chromedp.NodeReady),
			chromedp.Sleep(c.timeBetweenSteps),
		)
	}

	return chromedp.Run(ctx)
}

// exportaPlanilha clica no botão correto para exportar para excel, espera um tempo para download renomeia o arquivo.
func (c crawler) exportaPlanilha(ctx context.Context, fName string, tipo string) error {
	pathPlan := "//*[@title='Enviar para Excel']"

	if tipo == "verbas"{
		pathPlan = "/html/body/div[5]/div/div[16]/div[1]/div[1]/div"
	} 

	chromedp.Run(ctx,
		// Clica no botão de download
		chromedp.Click(pathPlan, chromedp.BySearch, chromedp.NodeVisible),
		chromedp.Sleep(c.timeBetweenSteps),
	)

	if err := nomeiaDownload(c.output, fName); err != nil {
		return fmt.Errorf("erro renomeando arquivo (%s): %v", fName, err)
	}
	if _, err := os.Stat(fName); os.IsNotExist(err) {
		return fmt.Errorf("download do arquivo de %s não realizado", fName)
	}
	return nil
}

// nomeiaDownload dá um nome ao último arquivo modificado dentro do diretório
// passado como parâmetro nomeiaDownload dá pega um arquivo
func nomeiaDownload(output, fName string) error {
	// Identifica qual foi o ultimo arquivo
	files, err := os.ReadDir(output)
	if err != nil {
		return fmt.Errorf("erro lendo diretório %s: %v", output, err)
	}
	var newestFPath string
	var newestTime int64 = 0
	for _, f := range files {
		fPath := filepath.Join(output, f.Name())
		fi, err := os.Stat(fPath)
		if err != nil {
			return fmt.Errorf("erro obtendo informações sobre arquivo %s: %v", fPath, err)
		}
		currTime := fi.ModTime().Unix()
		if currTime > newestTime {
			newestTime = currTime
			newestFPath = fPath
		}
	}
	// Renomeia o ultimo arquivo modificado.
	if err := os.Rename(newestFPath, fName); err != nil {
		return fmt.Errorf("erro renomeando último arquivo modificado (%s)->(%s): %v", newestFPath, fName, err)
	}
	return nil
}
