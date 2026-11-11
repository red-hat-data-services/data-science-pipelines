import os
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import MagicMock

from infra_deployer import InfraDeployer
import yaml


class TestArgoLiteNamespace(unittest.TestCase):

    def test_deploy_argo_lite_uses_operator_namespace(self):
        for namespace in ('opendatahub', 'rhods'):
            with self.subTest(namespace=namespace), tempfile.TemporaryDirectory(
            ) as temp_dir:
                argo_lite_path = os.path.join(temp_dir, '.github', 'resources',
                                              'argo-lite')
                os.makedirs(argo_lite_path)
                kustomization_path = os.path.join(argo_lite_path,
                                                  'kustomization.yaml')
                with open(
                        kustomization_path, 'w',
                        encoding='utf-8') as kustomization_file:
                    yaml.safe_dump(
                        {
                            'apiVersion': 'kustomize.config.k8s.io/v1beta1',
                            'kind': 'Kustomization',
                            'namespace': 'opendatahub',
                            'resources': ['../../../config/argo'],
                        },
                        kustomization_file,
                        sort_keys=False,
                    )

                deployment_manager = MagicMock()
                deployer = InfraDeployer(
                    args=SimpleNamespace(),
                    deployment_manager=deployment_manager,
                    operator_repo_path=temp_dir,
                    deployment_namespace='kubeflow',
                    dspa_name='test-dspa',
                    operator_namespace=namespace,
                    temp_dir=temp_dir,
                )

                deployer.deploy_argo_lite()

                with open(
                        kustomization_path,
                        encoding='utf-8') as kustomization_file:
                    rendered_kustomization = yaml.safe_load(kustomization_file)
                self.assertEqual(namespace, rendered_kustomization['namespace'])
                deployment_manager.apply_resource.assert_called_once_with(
                    manifest_path=argo_lite_path,
                    namespace=namespace,
                    kustomize=True,
                    description='Argo Lite',
                )


if __name__ == '__main__':
    unittest.main()
