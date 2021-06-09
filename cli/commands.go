package cli

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/putdotio/go-putio"
)

type resultsChannel struct {
	message string
	err     error
}

// getFileNameByIDList is a helper function
func (a *Application) getFileNameByIDList(ids []string) map[int64]string {
	fileInfos := make(map[int64]string)
	idList := []int64{}
	for _, id := range ids {
		num, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		idList = append(idList, num)
	}

	fileInfoCh := make(chan putio.File)
	for _, fileID := range idList {
		go func(id int64) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			f, err := a.Client.Files.Get(ctx, id)
			if err == nil {
				fileInfoCh <- f
			}

		}(fileID)
	}

	for range ids {
		r := <-fileInfoCh
		fileInfos[r.ID] = strings.TrimSpace(r.Name)
	}

	return fileInfos
}

// deleteByIDList is a helper function for deleting files by multiple ids
func (a *Application) deleteByIDList(ids []int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Client.Files.Delete(ctx, ids...); err != nil {
		return err
	}
	return nil
}

// CommandDelete deletes given file ids
func (a *Application) CommandDelete() error {
	if len(FlagSetDelete.Args()) == 0 {
		return ErrMissingArgs
	}

	fileNamesMap := a.getFileNameByIDList(FlagSetDelete.Args())

	deleteIDS := []int64{}
	for id := range fileNamesMap {
		deleteIDS = append(deleteIDS, id)
	}

	if len(deleteIDS) > 0 {
		if err := a.deleteByIDList(deleteIDS); err != nil {
			return err
		}

		for fileID, fileName := range fileNamesMap {
			fmt.Fprintf(a.Out, "[%s] -> %s deleted\n", colorWhite(fileID), colorRed(fileName))
		}
	}
	return nil
}

// CommandUpload @wip - implement file upload soon!
func (a *Application) CommandUpload() error {
	if len(FlagSetUpload.Args()) == 0 {
		return ErrMissingArgs
	}
	subcommand := FlagSetUpload.Args()[0]

	switch subcommand {
	case "url":
		incomingURLs := FlagSetUpload.Args()[1:]

		if len(incomingURLs) == 0 {
			return ErrMissingArgs
		}

		needURLS := []string{}
		for _, u := range incomingURLs {
			url, err := url.Parse(u)
			if err == nil && url.Scheme != "" {
				needURLS = append(needURLS, url.String())
			}
		}
		transferResults := make(chan resultsChannel)

		for _, url := range needURLS {
			go func(url string) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				transfer, err := a.Client.Transfers.Add(ctx, url, 0, "")
				if err == nil {
					transferResults <- resultsChannel{transfer.StatusMessage, nil}
				} else {
					if putioErr, ok := err.(*putio.ErrorResponse); ok {
						transferResults <- resultsChannel{putioErr.Message + " for " + url, err}
					} else {
						transferResults <- resultsChannel{"error for " + url, err}
					}
				}
			}(url)
		}

		for _, url := range needURLS {
			r := <-transferResults
			if r.err == nil {
				fmt.Fprintf(a.Out, colorYellow("-> [%s] %s\n"), url, r.message)
			} else {
				fmt.Fprintf(a.Out, colorRed("-> [%s]\n"), r.message)
			}
		}
		close(transferResults)

	default:
		fmt.Println("wrong", subcommand)
	}
	return nil
}

// CommandMove @wip
func (a *Application) CommandMove() error {
	if len(FlagSetMove.Args()) < 2 {
		return ErrMissingArgs
	}

	fileNamesMap := a.getFileNameByIDList(FlagSetMove.Args())
	targetFolder, err := strconv.ParseInt(FlagSetMove.Args()[len(FlagSetMove.Args())-1], 10, 64)
	if err != nil {
		return err
	}

	sourceIDs := []int64{}
	for id := range fileNamesMap {
		if id != targetFolder {
			sourceIDs = append(sourceIDs, id)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Client.Files.Move(ctx, targetFolder, sourceIDs...); err != nil {
		return err
	}

	for fileID, fileName := range fileNamesMap {
		if fileID != targetFolder {
			fmt.Fprintf(a.Out, "%s moved to -> %s\n", colorWhite(fileName), colorBoldYellow(fileNamesMap[targetFolder]))
		}
	}
	return nil
}

// CommandList fetches give folder id
func (a *Application) CommandList() error {
	var rootDir int64

	if len(FlagSetList.Args()) > 0 {
		number, err := strconv.ParseInt(FlagSetList.Args()[0], 10, 64)
		if err == nil {
			rootDir = number
		}
	}

	// do delete operation is flag is set!
	if *FlagListDeleteByID > 0 {
		if err := a.deleteByIDList([]int64{*FlagListDeleteByID}); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	files, parent, err := a.Client.Files.List(ctx, rootDir)
	if err != nil {
		return err
	}

	filesLen := len(files)
	if filesLen > 0 {
		// stop everything and return id list in a single line
		if *FlagListShowOnlyID {
			for i, file := range files {
				fmt.Fprintf(a.Out, "%d", file.ID)
				if i < filesLen-1 {
					fmt.Fprintf(a.Out, " ")
				}
			}
			fmt.Fprintln(a.Out, "")
			return nil
		}

		// stop everything and return download urls as list
		if *FlagListShowDownloadURL {
			urlResults := make(chan resultsChannel)
			for _, file := range files {
				go func(fileID int64) {
					fileURL, err := a.Client.Files.URL(ctx, fileID, false)
					if err == nil {
						urlResults <- resultsChannel{fileURL, nil}
					} else {
						if putioErr, ok := err.(*putio.ErrorResponse); ok {
							urlResults <- resultsChannel{fmt.Sprintf("%s for %d", putioErr.Message, fileID), err}
						} else {
							urlResults <- resultsChannel{fmt.Sprintf("error for %d", fileID), err}
						}
					}
				}(file.ID)
			}

			// just display downloadable files...
			for range files {
				r := <-urlResults
				if r.err == nil {
					fmt.Fprintln(a.Out, r.message)
				}
			}
			close(urlResults)
			return nil
		}

		table := simpletable.New()
		table.SetStyle(simpletable.StyleDefault)
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Text: colorBoldYellow("ID")},
				{},
				{Text: colorBoldYellow("> " + parent.Name)},
				{Text: colorBoldYellow("Size")},
			},
		}

		for _, file := range files {
			fileID := colorBoldWhite(strconv.FormatInt(file.ID, 10))
			isDir := "---"
			fileName := file.Name
			fileSize := a.ByteCountSI(file.Size)

			if file.IsDir() {
				isDir = colorBoldBlue("dir")
				fileName = colorBoldBlue(file.Name)
				fileSize = colorBoldBlue(fileSize)
			}

			r := []*simpletable.Cell{
				{Text: fileID},
				{Text: isDir},
				{Text: fileName},
				{Align: simpletable.AlignRight, Text: fileSize},
			}
			table.Body.Cells = append(table.Body.Cells, r)
		}
		fmt.Fprintln(a.Out, table.String())
		return nil
	}

	fmt.Fprintf(a.Out, "sorry, folder: %s is empty\n", colorRed(parent.Name))
	return nil
}
