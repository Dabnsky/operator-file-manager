/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"
	"os"
	"path/filepath"

	filev1 "github.com/example/file-operator/api/v1"
	"github.com/go-logr/logr"
)

// FileReconciler reconciles a File object
type FileReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=file.example.com,resources=files,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=file.example.com,resources=files/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=file.example.com,resources=files/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the File object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
// Reconcile compares the actual state with the desired state and takes action to reconcile the difference
func (r *FileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("file", req.NamespacedName)

	// Fetch the File instance
	file := &filev1.File{}
	if err := r.Get(ctx, req.NamespacedName, file); err != nil {
		log.Error(err, "unable to fetch File")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Retrieve the directory path and the maximum number of files allowed from the spec
	directoryPath := file.Spec.DirectoryPath
	maxFiles := file.Spec.MaxFiles

	// Count the number of files in the directory
	numFiles, err := countFiles(directoryPath)
	if err != nil {
		log.Error(err, "unable to count files")
		return ctrl.Result{}, err
	}

	// Compare the actual number of files with the maximum allowed
	if numFiles != maxFiles {
		// Take action to manage the files (e.g., delete old files, raise an alert, etc.)
		err := manageFiles(directoryPath, numFiles, maxFiles)
		if err != nil {
			log.Error(err, "unable to manage files")
			return ctrl.Result{}, err
		}

		// Update the status of the File instance to reflect the current state
		file.Status.NumFiles = numFiles
		if err := r.Status().Update(ctx, file); err != nil {
			log.Error(err, "unable to update File status")
			return ctrl.Result{}, err
		}

		log.Info("managed files successfully")
	}

	return ctrl.Result{}, nil
}

// countFiles counts the number of files in the specified directory
func countFiles(directoryPath string) (int, error) {

	var fileCount int = 0

	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return 0, err
	}

	fmt.Printf("Total files in directory %s: %d\n", directoryPath, fileCount)

	return fileCount, nil

}

// manageFiles takes action to manage the files in the directory
func manageFiles(directoryPath string, numFiles int, maxFiles int) error {
	// Implement logic to manage files (e.g., delete old files, raise an alert, etc.)

	if numFiles > maxFiles {
		numToDelete := numFiles - maxFiles
		for i := 0; i < numToDelete; i++ {
			file, err := os.Create("filename")
			if err != nil {
				return err
			}
			defer file.Close()
		}
	}

	if numFiles < maxFiles {
		numToCreate := maxFiles - numFiles
		for i := 0; i < numToCreate; i++ {
			file, err := os.Create("filename")
			if err != nil {
				return err
			}
			defer file.Close()
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&filev1.File{}).
		Complete(r)
}
