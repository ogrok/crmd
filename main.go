package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/ogrok/crmd/models"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	Directory   = "/.crmd"
	StorageFile = "/reminders.json"

	RecurrenceDaily = "daily"
	RecurrenceWeekly = "weekly"
	RecurrenceMonthly = "monthly"
	RecurrenceQuarterly = "quarterly"
	RecurrenceYearly = "yearly"
)

func main() {
	// first init array of valid recurrences
	validRecurrences := []string{
		RecurrenceDaily,
		RecurrenceWeekly,
		RecurrenceMonthly,
		RecurrenceQuarterly,
		RecurrenceYearly,
	}

	// list possible flags to be inserted into here
	var opts struct {
		Date     string `short:"d" description:"Date of new reminder"`
		Time     string `short:"t" description:"Optional time of new reminder"`
		Recur    string `short:"r" description:"Recurrence schedule for new reminder"`
		Complete int    `short:"c" description:"ID of reminder to mark complete"`
		Delete   int    `short:"x" description:"ID of reminder to delete"`
		ListAll  bool   `short:"a" description:"List all reminders that currently exist"`
	}

	descArray, _ := flags.Parse(&opts)
	description := strings.Join(descArray, " ")

	hasDate := len(opts.Date) > 0
	hasTime := len(opts.Time) > 0
	hasRecur := len(opts.Recur) > 0
	hasDescription := len(descArray) > 0
	hasComplete := opts.Complete != 0
	hasDelete := opts.Delete != 0
	hasListAll := opts.ListAll

	// if recurrence is passed, validate it
	if hasRecur {
		validRecur := false
		for _, v := range validRecurrences {
			if v == opts.Recur {
				validRecur = true
				break
			}
		}

		if !validRecur {
			fmt.Println("invalid recurrence schedule: "+opts.Recur+"\n"+
				"valid schedules: daily, weekly, monthly, quarterly, yearly")
			os.Exit(0)
		}
	}

	// some commands are only valid with no extraneous things
	validComplete := true
	if hasDate || hasTime || hasRecur || hasDescription || hasDelete || hasListAll {
		validComplete = false
	}
	validDelete := true
	if hasDate || hasTime || hasRecur || hasDescription || hasComplete || hasListAll {
		validDelete = false
	}
	validListAll := true
	if hasDate || hasTime || hasRecur || hasDescription || hasComplete || hasDelete {
		validListAll = false
	}

	if hasComplete {
		if !validComplete {
			fmt.Println("can only use -c flag by itself")
			os.Exit(0)
		}

		result, err := completeReminder(opts.Complete, true)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		fmt.Println(result)
		os.Exit(0)
	}

	if hasDelete {
		if !validDelete {
			fmt.Println("can only use -x flag by itself")
			os.Exit(0)
		}

		result, err := completeReminder(opts.Delete, false)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		fmt.Println(result)
		os.Exit(0)
	}

	if hasListAll {
		if !validListAll {
			fmt.Println("can only use -a flag by itself")
			os.Exit(0)
		}

		checkReminders(true)
		os.Exit(0)
	}

	if hasDescription {
		if !hasDate {
			fmt.Println("cannot create reminder: no date provided")
			os.Exit(0)
		}

		timestamp, err := toUnixDate(opts.Date, opts.Time)
		if err != nil {
			 fmt.Println(err.Error())
		}

		result, err := createReminder(description, timestamp, opts.Recur)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}

		fmt.Println(result)
		os.Exit(0)
	} else {
		if hasDate || hasTime || hasRecur {
			fmt.Println("description not found; ignoring flags")
		}

		checkReminders(false)
	}
}

// ---------------
// --- FILE I/O --
// ---------------

func createReminder(description string, datetime int64, recurrence string) (string, error) {
	reminders, err := loadRemindersFile()
	if err != nil {
		return "", err
	}

	// get lowest ID not in use
	lowest := 1
	for {
		validated := true
		for _, v := range reminders {
			if v.ID == lowest {
				validated = false
				break
			}
		}
		if validated {
			break
		}
		lowest++
	}

	reminders = append(reminders, models.Reminder{
		ID:          lowest,
		Description: description,
		Recurrence:  recurrence,
		Timestamp:   datetime,
	})

	err = persist(reminders)
	if err != nil {
		return "", err
	}

	return "Created reminder " + strconv.Itoa(lowest) + ".", nil
}

// checkReminders checks the file either for all reminders, or those with
// timestamps prior to the present moment, and prints those for the user.
func checkReminders(listAll bool) {
	all, err := loadRemindersFile()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if listAll && len(all) == 0 {
		fmt.Println("no reminders exist")
		return
	}

	for _, v := range all {
		if listAll || v.Timestamp <= time.Now().Unix() {


			s := "Reminder! #"+strconv.Itoa(v.ID)+" - "+time.Unix(v.Timestamp, 0).
				Format("2 Jan 2006")+ ": "+ v.Description

			if len(v.Recurrence) > 0 {
				s += " - recurs "+v.Recurrence
			}

			fmt.Println(s)
		}
	}
}

func completeReminder(id int, allowRecurrence bool) (string, error) {
	reminders, err := loadRemindersFile()
	if err != nil {
		return "", err
	}

	i := -1
	for k, v := range reminders {
		if v.ID == id {
			i = k
			break
		}
	}

	if i < 0 {
		return "", errors.New("reminder " + strconv.Itoa(id) + " not found")
	}

	willRecur := false
	if !allowRecurrence || len(reminders[i].Recurrence) == 0 {
		reminders[i] = reminders[len(reminders)-1]
		reminders = reminders[0:len(reminders)-1]
	} else {
		next, err := nextRecurrence(reminders[i])
		if err != nil {
			return "", err
		}
		reminders[i].Timestamp = next
		willRecur = true
	}

	err = persist(reminders)
	if err != nil {
		 return "", err
	}

	if willRecur {
		next := time.Unix(reminders[i].Timestamp, 0)
		return "Resolved reminder "+strconv.Itoa(id)+". Next occurrence: "+next.String(), nil
	} else if allowRecurrence {
		return "Resolved reminder "+strconv.Itoa(id)+".", nil
	} else {
		return "Deleted reminder "+strconv.Itoa(id)+".", nil
	}
}

func loadRemindersFile() ([]models.Reminder, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	directory := home + Directory
	fileLocation := directory + StorageFile

	if !fileExists(fileLocation) {
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			_ = os.Mkdir(directory, 0777)
		}

		newFile, err := os.Create(fileLocation)
		if err != nil {
			return nil, err
		}

		defer newFile.Close()

		_, err = newFile.WriteString("[]\n")
		if err != nil {
			return nil, err
		}
	}

	jsonFile, err := os.Open(fileLocation)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var output []models.Reminder
	json.Unmarshal(bytes, &output)

	return output, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// persist sorts the input slice of reminders and saves over
// the existing set. This means this function must be passed an
// exhaustive list of reminders that should exist. An error
// guarantees a write did not take place.
func persist(input []models.Reminder) error {
	sort.Slice(input, func(i, j int) bool {
		return input[i].Timestamp < input[j].Timestamp
	})

	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	directory := home + Directory

	file, err := os.Create(directory + StorageFile)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(bytes)
	return err
}

// ---------------
// --- HELPERS ---
// ---------------

func toUnixDate(d string, t string) (int64, error) {
	// try to parse with both date and time
	result, err := time.ParseInLocation("2006-01-02 15:04", d+" "+t, time.Local)

	if err != nil {
		if len(t) > 0 {
			fmt.Println("note: could not parse time component: "+t)
		}
		result, err = time.ParseInLocation("2006-01-02", d, time.Local)
		if err != nil {
			return 0, errors.New("could not parse time: "+ d+" "+t)
		}
	}

	return result.Unix(), nil
}

// nextRecurrence calculates the next time that a reminder should recur,
// relative to now, if recurrence is present. If not, function errors.
func nextRecurrence(input models.Reminder) (int64, error) {
	date := time.Unix(input.Timestamp, 0)
	now := time.Now()

	switch input.Recurrence {
	case RecurrenceDaily:
		date = date.AddDate(0, 0, 1)
		for now.After(date) {
			date = date.AddDate(0, 0, 1)
		}
	case RecurrenceWeekly:
		date = date.AddDate(0, 0, 7)
		for now.After(date) {
			date = date.AddDate(0, 0, 7)
		}
	case RecurrenceMonthly:
		date = date.AddDate(0, 1, 0)
		for now.After(date) {
			date = date.AddDate(0, 1, 0)
		}
	case RecurrenceQuarterly:
		date = date.AddDate(0, 3, 0)
		for now.After(date) {
			date = date.AddDate(0, 3, 0)
		}
	case RecurrenceYearly:
		date = date.AddDate(1, 0, 0)
		for now.After(date) {
			date = date.AddDate(1, 0, 0)
		}
	default:
		return -1, errors.New("could not find next recurrence date; " +
			"blank or invalid recurrence schedule")
	}

	return date.Unix(), nil
}
